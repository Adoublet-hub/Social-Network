import { useState } from "react";
import { motion } from "framer-motion";
import { IconX } from "@tabler/icons-react";

export default function CreatePost({ closeModal, refreshPosts }) {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [visibility, setVisibility] = useState("public");
  const [image, setImage] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleImageChange = (e) => {
    setImage(e.target.files[0]);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    const formData = new FormData();
    formData.append("title", title);
    formData.append("content", content);
    formData.append("visibility", visibility);
    if (image) formData.append("image", image);

    try {
      const response = await fetch("http://127.0.0.1:8079/create_post", {
        method: "POST",
        headers: { Authorization: `Bearer ${localStorage.getItem("authToken")}` },
        body: formData,
      });

      if (!response.ok) {
        const errorResponse = await response.json();
        setError(errorResponse.message || "Une erreur est survenue.");
      } else {
        const newPost = await response.json();
        refreshPosts(newPost); 
        closeModal && closeModal();
      }
    } catch (err) {
      console.error("Erreur lors de la soumission :", err);
      setError("Impossible de soumettre le post.");
    } finally {
      setLoading(false);
    }
  };
  return (
    <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 p-4 z-50">
      <div className="max-w-lg w-full bg-gray-900 shadow-lg rounded-lg p-6 border border-gray-700 relative">
        
        {/* Bouton Fermer (Annuler) */}
        <button onClick={closeModal} className="absolute top-4 right-4 text-gray-400 hover:text-red-500">
          <IconX className="w-6 h-6" />
        </button>

        <motion.h2
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-2xl font-bold text-center text-cyan-400 mb-4"
        >
          Nouveau Post
        </motion.h2>

        {error && <p className="mb-4 text-center text-red-500">{error}</p>}

        <form onSubmit={handleSubmit} encType="multipart/form-data" className="space-y-4">
          <div>
            <label className="block font-semibold mb-1 text-sm text-cyan-300">Titre</label>
            <input
              type="text"
              className="w-full p-2 rounded-lg bg-gray-800 border border-gray-700 text-gray-300"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="block font-semibold mb-1 text-sm text-cyan-300">Contenu</label>
            <textarea
              className="w-full p-2 rounded-lg bg-gray-800 border border-gray-700 text-gray-300"
              rows="3"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="block font-semibold mb-1 text-sm text-cyan-300">Image (optionnelle)</label>
            <input type="file" accept="image/*" onChange={handleImageChange} className="w-full text-sm" />
          </div>

          {/* Boutons Publier et Annuler */}
          <div className="flex justify-between mt-4">
            <button
              type="button"
              onClick={closeModal}
              className="w-1/3 bg-gray-700 text-white py-2 rounded-lg hover:bg-gray-600"
            >
              Annuler
            </button>
            <motion.button
              type="submit"
              className="w-2/3 bg-cyan-600 text-white py-2 rounded-lg hover:bg-cyan-500"
              disabled={loading}
            >
              {loading ? "En cours..." : "Publier"}
            </motion.button>
          </div>
        </form>
      </div>
    </div>
  );
}

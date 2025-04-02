import { useState } from "react";

export default function CreateGroupPost({ groupId, onPostCreated }) {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch("http://127.0.0.1:8079/create_post_group", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: JSON.stringify({ group_id: groupId, title, content }),
      });

      if (response.ok) {
        setTitle("");
        setContent("");
        onPostCreated();
      }
    } catch (error) {
      console.error("Erreur lors de la création du post :", error);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="mb-4">
      <input
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Titre du post"
        className="w-full p-2 mb-2 bg-gray-700 rounded"
        required
      />
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="Contenu du post"
        className="w-full p-2 bg-gray-700 rounded"
        required
      ></textarea>
      <button type="submit" className="bg-cyan-600 text-white p-2 rounded mt-2">Créer le post</button>
    </form>
  );
}

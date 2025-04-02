import { useState } from "react";
import { useRouter } from "next/router";
import { motion } from "framer-motion";
import { IconArrowLeft } from "@tabler/icons-react";

export default function UpdateProfilePage() {
  const [profile, setProfile] = useState({
    firstName: "",
    lastName: "",
    email: "",
    gender: "",
    avatar: "",
    bio: "",
    phoneNumber: "",
    address: "",
    isPrivate: false,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [successMessage, setSuccessMessage] = useState(null);
  const [imageLoading, setImageLoading] = useState(false);
  const router = useRouter();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setProfile((prevProfile) => ({ ...prevProfile, [name]: value }));
  };

  const handleToggleVisibility = () => {
    setProfile((prevProfile) => ({ ...prevProfile, isPrivate: !prevProfile.isPrivate }));
  };

  const handleImageChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      setImageLoading(true);
      reader.onloadend = () => {
        setProfile((prevProfile) => ({ ...prevProfile, avatar: reader.result }));
        setImageLoading(false);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch("http://127.0.0.1:8079/update_profile", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: JSON.stringify(profile),
      });

      const responseBody = await response.clone().json().catch(() => response.text());
      
      if (!response.ok) {
        setError(responseBody.message || "Échec de la mise à jour.");
      } else {
        setSuccessMessage("Profil mis à jour avec succès !");
        setTimeout(() => router.push("/profil/myprofil"), 2000);
      }
    } catch (error) {
      setError("Erreur lors de la mise à jour du profil.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-gray-200 flex flex-col items-center p-6">
      <header className="w-full max-w-3xl bg-gray-800 shadow-md flex items-center p-3 rounded-lg">
        <motion.button
          onClick={() => router.push("/profil/myprofil")}
          whileHover={{ scale: 1.05 }}
          className="flex items-center text-cyan-400 hover:text-cyan-300 transition duration-200"
        >
          <IconArrowLeft className="h-5 w-5 mr-2" />
          Retour au profil
        </motion.button>
      </header>

      <div className="max-w-3xl w-full my-8 bg-gray-800 rounded-xl shadow-lg p-8">
        <motion.h2
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-3xl font-semibold text-center text-cyan-400 mb-6"
        >
          Modifier le Profil
        </motion.h2>

        {error && <p className="mb-4 text-center text-red-500 text-sm">{error}</p>}
        {successMessage && <p className="mb-4 text-center text-green-500 text-sm">{successMessage}</p>}

        <form onSubmit={handleSubmit} className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="md:col-span-2">
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Image de profil</label>
            <input type="file" accept="image/*" onChange={handleImageChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200" />
            {profile.avatar && <img src={profile.avatar} alt="Profile Preview" className="mt-4 h-24 w-24 rounded-full object-cover mx-auto" />}
          </div>
          
          <div>
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Prénom</label>
            <input type="text" name="firstName" value={profile.firstName} onChange={handleChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200" />
          </div>
          <div>
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Nom</label>
            <input type="text" name="lastName" value={profile.lastName} onChange={handleChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200" />
          </div>
          <div>
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Email</label>
            <input type="email" name="email" value={profile.email} onChange={handleChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200" />
          </div>
          <div>
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Genre</label>
            <select name="gender" value={profile.gender} onChange={handleChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200">
              <option value="">Sélectionnez</option>
              <option value="Homme">Homme</option>
              <option value="Femme">Femme</option>
            </select>
          </div>
          <div className="md:col-span-2">
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Bio</label>
            <textarea name="bio" value={profile.bio} onChange={handleChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200" rows="3" maxLength="500" />
          </div>
          <div>
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Téléphone</label>
            <input type="text" name="phoneNumber" value={profile.phoneNumber} onChange={handleChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200" />
          </div>
          <div>
            <label className="block font-semibold mb-2 text-sm text-cyan-300">Adresse</label>
            <input type="text" name="address" value={profile.address} onChange={handleChange} className="w-full p-3 rounded-lg bg-gray-900 text-gray-200" />
          </div>

          <motion.button type="submit" className="md:col-span-2 w-full p-3 bg-cyan-600 text-white font-semibold rounded-lg hover:bg-cyan-500 transition duration-300" disabled={loading}>
            {loading ? "En cours..." : "Mettre à jour"}
          </motion.button>
        </form>
      </div>
    </div>
  );
}

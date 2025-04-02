import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { apiRequest } from "../profil/SearchBar"; 
import { motion } from "framer-motion";
import {
  IconArrowLeft,
  IconHeart,
  IconMessageCircle,
} from "@tabler/icons-react";

export default function ViewProfile() {
  const router = useRouter();
  const { userId } = router.query;
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!userId) return;

    const fetchProfile = async () => {
      try {
        const data = await apiRequest(`/viewprofil/${userId}`);
        console.log("Profil récupéré :", data);
        setProfile(data);
      } catch (error) {
        setError("Ce compte est privé ou n'existe pas.");
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, [userId]);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen bg-gray-900">
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.5 }}
          className="text-lg font-bold text-gray-500"
        >
          Chargement...
        </motion.div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-gray-200">
      <header className="w-full px-4 py-3 bg-gray-800 shadow-md flex items-center justify-between">
        <motion.button
          onClick={() => router.push("/")}
          whileHover={{ scale: 1.05 }}
          className="flex items-center text-cyan-400 hover:text-cyan-300"
        >
          <IconArrowLeft className="h-5 w-5 mr-2" />
          Retour à l'accueil
        </motion.button>
      </header>

      <main className="max-w-4xl mx-auto my-8 px-4 sm:px-8 py-6 bg-gray-800 rounded-xl">
        {error ? (
          <div className="text-center text-red-400">{error}</div>
        ) : (
          <section>
            {/* Profil */}
            <div className="flex items-center space-x-4">
              <motion.img
                src={profile?.avatar || "/avatar.png"}
                alt="Avatar"
                className="w-24 h-24 sm:w-28 sm:h-28 rounded-full border-4 border-cyan-500 object-cover shadow-lg"
                whileHover={{ scale: 1.05 }}
              />
              <h2 className="text-3xl font-bold text-cyan-400">{profile?.username}</h2>
            </div>

            {/* Informations du profil */}
            <div className="mt-8 space-y-6">
              {[
                { label: "Nom complet", value: `${profile?.first_name?.String || ""} ${profile?.last_name?.String || ""}` },
                { label: "Biographie", value: profile?.bio?.String || "Pas de biographie" },
                { label: "Confidentialité", value: profile?.is_private ? "Privé" : "Public" },
              ].map((item, idx) => (
                <div key={idx} className="p-4 bg-gray-700 rounded-lg shadow-md">
                  <h3 className="text-xl font-semibold text-cyan-300">{item.label} :</h3>
                  <p className="text-lg text-gray-300">{item.value}</p>
                </div>
              ))}

              {/* Liste des abonnés */}
              <div className="p-4 bg-gray-700 rounded-lg shadow-md">
                <h3 className="text-xl font-semibold text-cyan-300">Abonnés :</h3>
                {profile?.followers?.length > 0 ? (
                  <ul>
                    {profile.followers.map((follower) => (
                      <li key={follower.id} className="text-gray-300">{follower.username}</li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-gray-300">Aucun abonné.</p>
                )}
              </div>

              {/* Liste des abonnements */}
              <div className="p-4 bg-gray-700 rounded-lg shadow-md">
                <h3 className="text-xl font-semibold text-cyan-300">Abonnements :</h3>
                {profile?.following?.length > 0 ? (
                  <ul>
                    {profile.following.map((followed) => (
                      <li key={followed.id} className="text-gray-300">{followed.username}</li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-gray-300">Aucun abonnement.</p>
                )}
              </div>

              {/* Publications */}
              <div className="p-4 bg-gray-700 rounded-lg shadow-md">
                <h3 className="text-xl font-semibold text-cyan-300">Publications :</h3>
                {profile?.posts?.length > 0 ? (
                  <div className="space-y-4">
                    {profile.posts.map((post) => (
                      <div key={post.id} className="bg-gray-800 p-4 rounded-lg shadow-md">
                        <h4 className="text-lg font-semibold text-cyan-300">{post.title}</h4>
                        <p className="text-gray-300">{post.content}</p>
                        {post.image_path && (
                          <img
                            src={post.image_path.startsWith("http") ? post.image_path : `http://127.0.0.1:8079/uploads/${post.image_path.replace(/^image_path\//, "")}`}
                            alt="Post"
                            className="mt-2 rounded-lg shadow-md w-full"
                          />
                        )}
                        <p className="text-sm text-gray-400">Publié le {new Date(post.created_at).toLocaleDateString()}</p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-gray-300">Aucune publication.</p>
                )}
              </div>
            </div>
          </section>
        )}
      </main>
    </div>
  );
}

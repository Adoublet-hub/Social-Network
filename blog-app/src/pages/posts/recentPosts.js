import { useState, useEffect } from "react";
import Image from "next/image";
import { IconMessageCircle } from "@tabler/icons-react";
import LikeButton from "./LikeButton";
import CommentSection from "./CommentSection";

export default function RecentPosts({ newPost }) {  
  const [posts, setPosts] = useState([]);
  const [showComments, setShowComments] = useState({});

  // Récupérer les posts au montage
  useEffect(() => {
    fetchPosts();
  }, []);

  // Ajout du nouveau post en haut de la liste si `newPost` change
  useEffect(() => {
    if (newPost) {
      setPosts((prevPosts) => [newPost, ...prevPosts]); 
    }
  }, [newPost]); 

  // Récupérer les posts depuis l'API
  const fetchPosts = async () => {
    try {
      const response = await fetch("http://127.0.0.1:8079/recent_posts?page=1&limit=10", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        }
      });

      if (!response.ok) throw new Error("Échec de récupération des posts");

      const data = await response.json();
      setPosts(data);
    } catch (error) {
      console.error("Erreur lors du chargement des posts :", error);
    }
  };

  // Basculer l'affichage des commentaires
  const toggleComments = (postId) => {
    setShowComments((prev) => ({
      ...prev,
      [postId]: !prev[postId]
    }));
  };

  return (
    <div className="bg-gray-700 p-4 lg:p-6 rounded-lg shadow-md border border-gray-600">
      <h2 className="text-2xl font-bold text-gray-300 mb-4 border-b border-gray-600 pb-2">
        Vos Posts Récents
      </h2>
      <div className="space-y-4 lg:space-y-6">
        {posts.length === 0 ? (
          <p className="text-gray-400 text-center">Aucun post pour le moment.</p>
        ) : (
          posts.map((post, idx) => (
            <div
              key={post.id || idx}
              className="p-6 bg-gray-800 shadow-lg rounded-lg hover:shadow-xl transition duration-300 border border-gray-700"
            >
              {/* Informations de l'utilisateur */}
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-4">
                  <Image
                    src={post.image_profil?.String || "/avatar.png"}
                    alt="avatar"
                    width={50}
                    height={50}
                    className="rounded-full"
                  />
                  <div>
                    <h3 className="text-lg font-semibold text-gray-300">
                      {post.username}
                    </h3>
                    <p className="text-gray-400 text-sm">
                      Posté le {post.created_at ? new Date(post.created_at).toLocaleDateString("fr-FR") : "Date inconnue"} à {post.created_at ? new Date(post.created_at).toLocaleTimeString("fr-FR") : ""}
                    </p>
                  </div>
                </div>
              </div>

              {/* Contenu du post */}
              <div className="mb-4">
                <p className="text-gray-400">{post.content || "Contenu du post indisponible"}</p>
              </div>

              {/* Image du post */}
              {post.image_path && (
                <div className="mt-4">
                  <Image
                    src={post.image_path.startsWith("http") ? post.image_path : `http://127.0.0.1:8079/${post.image_path}`}
                    alt="Post Image"
                    width={500}
                    height={300}
                    className="rounded-lg object-cover"
                    unoptimized
                  />
                </div>
              )}

              {/* Actions : Like et Commentaires */}
              <div className="flex items-center mt-4 space-x-4">
                <LikeButton 
                  postId={post.id} 
                  initialLikes={post.total_likes || 0} 
                  likedByUser={post.liked_by_user} 
                />

                <button onClick={() => toggleComments(post.id)} className="flex items-center text-gray-400 hover:text-cyan-400 transition-colors">
                  <IconMessageCircle className="h-5 w-5 mr-1" /> Commenter <span className="ml-2">{post.comments || 0}</span>
                </button>
              </div>

              {/* Section Commentaires */}
              {showComments[post.id] && <CommentSection postId={post.id} />}
            </div>
          ))
        )}
      </div>
    </div>
  );
}

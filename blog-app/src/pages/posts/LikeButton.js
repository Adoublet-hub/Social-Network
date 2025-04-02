import { useState } from "react";
import { IconHeart } from "@tabler/icons-react";

export default function LikeButton({ postId, initialLikes, likedByUser }) {
  const [likes, setLikes] = useState(initialLikes);
  const [liked, setLiked] = useState(likedByUser);

  const handleLikeToggle = async () => {
    const action = liked ? "unlike_post" : "like_post";
    const postData = { post_id: postId };

    console.log("Données envoyées :", postData);


    try {
      const response = await fetch(`http://127.0.0.1:8079/${action}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: JSON.stringify({ post_id: postId }),
      });

      console.log(response)

      if (!response.ok) throw new Error("Échec de l'action like/unlike");

      // Mettre à jour l'interface utilisateur
      setLikes((prevLikes) => liked ? prevLikes - 1 : prevLikes + 1);
      setLiked(!liked);
    } catch (error) {
      console.error("Erreur lors de l'action like/unlike:", error.message || error);
    }
  };

  return (
    <button
      onClick={handleLikeToggle}
      className={`flex items-center ${liked ? "text-blue-500" : "text-gray-400"} hover:text-cyan-400 transition-colors`}
    >
      <IconHeart className="h-5 w-5 mr-1" />
      {liked ? "J'aime" : "J'aime"} <span className="ml-2">({likes})</span>
    </button>
  );
}

// LikeButtonComment.js
import { useState } from "react";
import { IconHeart } from "@tabler/icons-react";

export default function LikeButtonComment({ commentID, initialLikes, likedByUser, onLikeToggle }) {
  const [likes, setLikes] = useState(initialLikes);
  const [liked, setLiked] = useState(likedByUser);
  const [loading, setLoading] = useState(false);

  const handleLikeToggle = async () => {
    if (loading) return;
    setLoading(true);

    const action = liked ? "unlike_comment" : "like_comment";
    const commentData = { comment_id: commentID };

    try {
      const response = await fetch(`http://127.0.0.1:8079/${action}`, {
          method: "POST",
          headers: {
              "Content-Type": "application/json",
              "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
          },
          body: JSON.stringify(commentData),
      });
  
      console.log(response)
  
      if (!response.ok) throw new Error("Échec de l'action like/unlike");
  
      setLikes((prevLikes) => liked ? prevLikes - 1 : prevLikes + 1);
      setLiked(!liked);
  
      onLikeToggle(commentID, liked ? likes - 1 : likes + 1, !liked);
    } catch (error) {
        console.error("Erreur lors de l'action like/unlike comment:", error.message || error);
    } finally {
        setLoading(false);
    }
  
  };

  return (
    <button
      onClick={handleLikeToggle}
      disabled={loading}
      className={`flex items-center ${liked ? "text-blue-500" : "text-gray-400"} hover:text-cyan-400 transition-colors`}
    >
      <IconHeart className="h-5 w-5 mr-1" />
      {liked ? "J'aime" : "J'aime"} <span className="ml-2">({likes})</span>
    </button>
  );
}

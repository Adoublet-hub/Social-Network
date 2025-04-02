// CommentSection.js
import { useState, useEffect } from "react";
import Image from "next/image";
import LikeButtonComment from "./LikeButtonComment"; 

export default function CommentSection({ postId }) {
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState("");

  useEffect(() => {
    async function fetchComments() {
      try {
        console.log("Fetching comments for postId:", postId);
        const response = await fetch(`http://127.0.0.1:8079/list_comment?post_id=${postId}`, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
          },
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          console.error("Error response text:", errorText);
          throw new Error("Failed to fetch comments");
        }
        
        const data = await response.json();
        setComments(data.comments || []);
      } catch (error) {
        console.error("Error fetching comments:", error.message || error);
      }
    }

    fetchComments();
  }, [postId]);

  const handleLikeToggle = (commentId, updatedLikes, likedStatus) => {
    setComments((prevComments) =>
      prevComments.map((comment) =>
        comment.id === commentId
          ? { ...comment, total_likes: updatedLikes, liked_by_user: likedStatus }
          : comment
      )
    );
  };

  const handleNewCommentSubmit = async () => {
    try {
      const response = await fetch("http://127.0.0.1:8079/create_comment", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: JSON.stringify({ post_id: postId, content: newComment }),
      });

      if (!response.ok) {
        throw new Error("Failed to create comment");
      }

      const createdComment = await response.json();
      setComments((prevComments) => [...prevComments, createdComment]);
      setNewComment("");
    } catch (error) {
      console.error("Error creating comment:", error.message || error);
    }
  };

  return (
    <div className="mt-4">
      <h4 className="text-gray-400 text-sm mb-2">Commentaires</h4>
      <div className="space-y-4">
        {comments.map((comment) => (
          <div key={comment.id} className="flex items-start space-x-4">
            <Image
              src={comment.avatar?.String || "/avatar.png"}
              alt="avatar"
              width={40}
              height={40}
              className="rounded-full"
            />
            <div>
              <h5 className="text-gray-300 text-sm font-semibold">{comment.username}</h5>
              <p className="text-gray-400 text-sm">{comment.content}</p>
              <LikeButtonComment 
                commentID={comment.id} 
                initialLikes={comment.total_likes || 0} 
                likedByUser={comment.liked_by_user}
                onLikeToggle={handleLikeToggle}
              />
            </div>
          </div>
        ))}
      </div>

      <div className="mt-4">
        <input
          type="text"
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Ajouter un commentaire..."
          className="bg-gray-600 p-2 rounded-lg w-full text-gray-300"
        />
        <button
          onClick={handleNewCommentSubmit}
          className="mt-2 bg-cyan-500 text-white py-1 px-4 rounded-lg"
        >
          Publier
        </button>
      </div>
    </div>
  );
}

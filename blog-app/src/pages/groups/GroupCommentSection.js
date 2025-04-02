import { useState, useEffect } from "react";

export default function GroupCommentSection({ postId }) {
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState("");

  useEffect(() => {
    fetchComments();
  }, []);

  const fetchComments = async () => {
    try {
      const response = await fetch(`http://127.0.0.1:8079/list_comments_group?post_id=${postId}`, {
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
      console.log("Commentaires récupérés :", data);  // ✅ Debug
      setComments(Array.isArray(data) ? data : []);   // Assurez-vous que c'est un tableau
    } catch (error) {
      console.error("Erreur lors de la récupération des commentaires :", error);
    }
  };
  
  const handleCommentSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch("http://127.0.0.1:8079/create_comment_group", {
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
  
      const result = await response.json();
      console.log("Commentaire créé :", result);  // ✅ Debug
  
      if (result.comment) {
        setComments((prevComments) => [...prevComments, result.comment]);
      }
  
      setNewComment("");
    } catch (error) {
      console.error("Erreur lors de la création du commentaire :", error);
    }
  };
  

  return (
    <div className="mt-2">
        {comments.map((comment, index) => (
        <div key={comment.id || `${comment.createdAt}-${index}`} className="bg-gray-700 p-2 rounded mb-1">
            <p className="text-gray-300">{comment.content}</p>
            <span className="text-xs text-gray-400">par {comment.username || "Utilisateur inconnu"}</span>  
        </div>
        ))}

  
      <form onSubmit={handleCommentSubmit}>
        <input
          type="text"
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Ajouter un commentaire"
          className="w-full p-1 bg-gray-700 rounded mt-2"
          required
        />
        <button type="submit" className="bg-cyan-500 text-white p-1 rounded mt-1">Commenter</button>
      </form>
    </div>
  );
  
}

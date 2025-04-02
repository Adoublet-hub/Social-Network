import { useState, useEffect } from "react";
import CreateGroupPost from "./CreateGroupPost";
import GroupCommentSection from "./GroupCommentSection";

export default function GroupPosts({ groupId }) {
  const [posts, setPosts] = useState([]); 
  const [error, setError] = useState(null);

  const fetchGroupPosts = async () => {
    try {
      const response = await fetch(`http://127.0.0.1:8079/list_post_group?group_id=${groupId}`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });

      if (!response.ok) throw new Error("Erreur lors de la récupération des posts.");

      const data = await response.json();
      setPosts(Array.isArray(data) ? data : []); 
    } catch (err) {
      console.error(err);
      setError("Impossible de charger les posts.");
    }
  };

  useEffect(() => {
    if (groupId) fetchGroupPosts();
  }, [groupId]);

  return (
    <div className="p-4 bg-gray-800 rounded-lg">
      <h2 className="text-xl text-white mb-4">Posts du Groupe</h2>

      <CreateGroupPost groupId={groupId} onPostCreated={fetchGroupPosts} />

      {error && <p className="text-red-400">{error}</p>}

      {posts?.length > 0 ? (
        posts.map((post) => (
            <div key={post.id} className="bg-gray-700 p-4 rounded-lg mb-4 shadow-md">
            <h3 className="text-lg text-white">{post.title}</h3>
            <p className="text-gray-400">{post.content}</p>
            <p className="text-sm text-cyan-400 mt-1">Posté par : {post.username}</p> {/* ✅ Affichage du username */}

            {/* Section des commentaires */}
            <GroupCommentSection postId={post.id} />
            </div>
        ))
        ) : (
        <p className="text-gray-400">Aucun post disponible.</p>
        )}


    </div>
  );
}

import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import { SidebarDemo } from "./ui/sidebar";
import RecentPosts from "./posts/recentPosts";
import { fetchNotifications, handleFollowAction, markNotificationAsRead } from "./profil/follow";
import { handleCreateGroup } from "./groups/ApiGroupjs";
import CreatePost from "./posts/create";
import Notifications from "./utils/notifComponent";
import { motion } from "framer-motion";
import {
  IconPlus,
  IconBell,
  IconMessage,
  IconUsersGroup,
} from "@tabler/icons-react";

export default function HomePage() {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [posts, setPosts] = useState([]);
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [showNotifications, setShowNotifications] = useState(false);

  // États pour les modals
  const [isGroupModalOpen, setIsGroupModalOpen] = useState(false);
  const [isPostModalOpen, setIsPostModalOpen] = useState(false);
  const [groupName, setGroupName] = useState("");
  const [groupDescription, setGroupDescription] = useState("");
  const [loadingGroup, setLoadingGroup] = useState(false); // Chargement spécifique pour le groupe

  // Récupération des notifications
  const loadNotifications = async () => {
    try {
      const data = await fetchNotifications();
      if (!Array.isArray(data)) throw new Error("Format inattendu des notifications.");
      setNotifications(data);
      setUnreadCount(data.filter((notif) => !notif.read).length);
    } catch (error) {
      console.error("Échec du chargement des notifications :", error);
      setNotifications([]);
      setUnreadCount(0);
    }
  };

  // Récupération des posts
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

  // Vérification de l'authentification et chargement des données
  useEffect(() => {
    const token = localStorage.getItem("authToken");
    if (!token) {
      router.replace("/login");
    } else {
      setIsAuthenticated(true);
      fetchPosts();
      loadNotifications();
    }
    setLoading(false);
  }, [router]);

  if (!isAuthenticated) return null;

  // Gérer une action sur une notification
  const handleNotificationAction = async (action, id) => {
    try {
      if (action === "markAsRead") {
        await markNotificationAsRead(id);
        setNotifications((prev) =>
          prev.map((notif) => (notif.id === id ? { ...notif, read: true } : notif))
        );
        setUnreadCount((prev) => prev - 1);
      } else {
        await handleFollowAction(action, id);
        setNotifications((prev) => prev.filter((notif) => notif.id !== id));
        setUnreadCount((prev) => prev - 1);
      }
    } catch (error) {
      console.error(`Erreur lors de l'action ${action} sur la notification ${id} :`, error);
    }
  };

  // Création de groupe
  const handleCreateGroupAction = async (e) => {
    e.preventDefault();
    setLoadingGroup(true);

    const groupData = { name: groupName, description: groupDescription };

    try {
      await handleCreateGroup(groupData);
      setGroupName("");
      setGroupDescription("");
      setIsGroupModalOpen(false);
      console.log("Groupe créé avec succès");
    } catch (error) {
      console.error("Erreur lors de la création du groupe :", error);
      alert("Échec de la création du groupe.");
    } finally {
      setLoadingGroup(false);
    }
  };

  // Rafraîchir les posts après création d'un nouveau post
  const refreshPosts = (newPost) => {
    setPosts((prevPosts) => [newPost, ...prevPosts]);
  };

  // Navigation
  const handleMessagesClick = () => router.push("/messages");
  const handleListGroupClick = () => router.push("/groups/GroupList");
return (
  <div className="flex flex-col lg:flex-row min-h-screen bg-gray-900 text-gray-300">
    {/* Sidebar */}
    <SidebarDemo />

    {/* Main Content */}
    <div className="flex-1 p-4 lg:p-8 bg-gray-800 shadow-lg rounded-lg m-5 relative">
      {/* Header */}
      <header className="fixed top-4 right-4 flex space-x-4 z-50 bg-gray-800 p-2 rounded-lg shadow-md">
        <button
          onClick={() => setShowNotifications(!showNotifications)}
          className="text-white relative hover:text-cyan-400 transition-colors"
        >
          <IconBell className="h-6 w-6 sm:h-8 sm:w-8" />
          {unreadCount > 0 && (
            <span className="absolute top-0 right-0 bg-red-600 text-xs rounded-full px-1">
              {unreadCount}
            </span>
          )}
        </button>
        <button
          onClick={handleMessagesClick}
          className="text-white relative hover:text-cyan-400 transition-colors"
        >
          <IconMessage className="h-6 w-6 sm:h-8 sm:w-8" />
        </button>
        <button
          onClick={handleListGroupClick}
          className="text-white hover:text-cyan-400 transition-colors"
        >
          <IconUsersGroup className="h-6 w-6 sm:h-8 sm:w-8" />
        </button>
      </header>

      {/* Notifications */}
      {showNotifications && (
        <Notifications
          notifications={notifications}
          setNotifications={setNotifications}  
          handleAccept={(id) => handleNotificationAction("accept", id)}
          handleRefuse={(id) => handleNotificationAction("decline", id)}
          handleMarkAsRead={(id) => handleNotificationAction("markAsRead", id)}
        />
      )}


      {/* Welcome Message */}
      <section className="mb-8 text-center border-b border-gray-700 pb-4">
        <h1 className="text-3xl font-bold text-white mb-2">Bienvenue sur votre espace !</h1>
        <p className="text-gray-400">
          Découvrez les événements, partagez vos pensées et connectez-vous avec le réseau.
        </p>
      </section>

      {/* Quick Action Cards */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="p-6 bg-green-600 text-white rounded-lg shadow-md hover:bg-green-500 transition duration-300 flex items-center cursor-pointer transform hover:scale-105"
          onClick={() => setIsPostModalOpen(true)}
        >
          <IconPlus className="h-8 w-8 mr-3" />
          <div>
            <h3 className="text-xl font-bold mb-2">Créer un Post</h3>
            <p>Partagez vos pensées avec le réseau.</p>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7 }}
          className="p-6 bg-purple-600 text-white rounded-lg shadow-md hover:bg-purple-500 transition duration-300 flex items-center cursor-pointer transform hover:scale-105"
          onClick={() => setIsGroupModalOpen(true)}
        >
          <IconPlus className="h-8 w-8 mr-3" />
          <div>
            <h3 className="text-xl font-bold mb-2">Créer Groupe</h3>
            <p>Créer un groupe pour votre communauté.</p>
          </div>
        </motion.div>
      </section>

      {/* Modal de création de post */}
      {isPostModalOpen && (
        <div className="fixed inset-0 bg-gray-900 bg-opacity-75 flex justify-center items-center">
          <div className="bg-gray-800 p-6 rounded-lg shadow-lg w-1/2">
            <CreatePost closeModal={() => setIsPostModalOpen(false)} refreshPosts={refreshPosts} />
          </div>
        </div>
      )}

      {/* Modal de création de groupe */}
      {isGroupModalOpen && (
        <div className="fixed inset-0 bg-gray-900 bg-opacity-75 flex justify-center items-center">
          <form
            onSubmit={handleCreateGroupAction}
            className="bg-gray-800 p-6 rounded-lg shadow-lg"
          >
            <h2 className="text-xl font-bold text-white mb-4">Créer un Groupe</h2>
            <div className="mb-4">
              <label className="block text-gray-400 mb-2">Nom du Groupe</label>
              <input
                type="text"
                value={groupName}
                onChange={(e) => setGroupName(e.target.value)}
                className="w-full p-2 bg-gray-700 rounded"
                required
              />
            </div>
            <div className="mb-4">
              <label className="block text-gray-400 mb-2">Description</label>
              <textarea
                value={groupDescription}
                onChange={(e) => setGroupDescription(e.target.value)}
                className="w-full p-2 bg-gray-700 rounded"
                required
              />
            </div>
            <div className="flex justify-end space-x-4">
              <button
                type="button"
                onClick={() => setIsGroupModalOpen(false)}
                className="bg-gray-500 px-4 py-2 rounded text-white hover:bg-gray-600"
              >
                Annuler
              </button>
              <button
                type="submit"
                className="bg-blue-600 px-4 py-2 rounded text-white hover:bg-blue-700"
              >
                Créer
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Affichage des posts */}
      <RecentPosts posts={posts} />
    </div>
  </div>
);
}  



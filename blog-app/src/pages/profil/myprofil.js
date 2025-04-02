import { useState, useEffect } from "react";
import Image from "next/image";
import { useRouter } from "next/router";
import { motion } from "framer-motion";
import { SearchBar, UserCard, apiRequest } from "./SearchBar"; 
import { followUser, unfollowUser, acceptFollowRequest, declineFollowRequest, searchUsers, fetchUsers } from "./follow";

import {
  IconEdit,
  IconArrowLeft,
  IconHeart,
  IconMessageCircle,
} from "@tabler/icons-react";

export default function ProfilePage() {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [followRequests, setFollowRequests] = useState([]);
  const [loadingRequests, setLoadingRequests] = useState(false);


  const [showUsersList, setShowUsersList] = useState(false);
  const [users, setUsers] = useState([]);
  const [filteredUsers, setFilteredUsers] = useState([]);
  const [pagination, setPagination] = useState({ page: 1, limit: 10 });
  const [usersLoading, setUsersLoading] = useState(false);

  const router = useRouter();

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const data = await apiRequest("/myprofil");
  
        console.log("Données du profil récupérées :", data); 
  
        setProfile(data);
        console.log("data: ", data)
      } catch (error) {
        setError("Impossible de charger le profil.");
      } finally {
        setLoading(false);
      }
    };
  
    fetchProfile();
  }, []);
  
  const fetchFollowRequests = async () => {
    setLoadingRequests(true);
    try {
      const response = await apiRequest("/follow_requests"); 
      setFollowRequests(response.data || []);
    } catch (error) {
      console.error("Erreur lors de la récupération des demandes de suivi :", error);
    } finally {
      setLoadingRequests(false);
    }
  };

  const handleAcceptFollow = async (requestID) => {
    try {
      await acceptFollowRequest(requestID);
      setFollowRequests(followRequests.filter((req) => req.id !== requestID)); 
    } catch (error) {
      console.error("Erreur lors de l'acceptation de la demande de suivi :", error);
    }
  };
  
  const handleDeclineFollow = async (requestID) => {
    try {
      await declineFollowRequest(requestID);
      setFollowRequests(followRequests.filter((req) => req.id !== requestID)); 
    } catch (error) {
      console.error("Erreur lors du refus de la demande de suivi :", error);
    }
  };
  
  
  
  const fetchUsers = async () => {
    setUsersLoading(true);
    try {
      const data = await apiRequest(`/list_users?page=${pagination.page}&limit=${pagination.limit}`);
      console.log("🔍 Données des utilisateurs récupérées :", data.results);
  
      if (!data.results) {
        setUsers([]);
        setFilteredUsers([]);
        return;
      }
  
      const updatedUsers = data.results.map((user) => ({
        ...user,
        isRequestPending: user.isRequestPending || false,
      }));
  
      setUsers(updatedUsers);
      setFilteredUsers(updatedUsers);
    } catch (error) {
      console.error("Erreur lors de la récupération des utilisateurs :", error);
    } finally {
      setUsersLoading(false);
    }
  };
  
  
  const toggleFollow = async (userID) => {
    try {
      console.log("🔄 ID utilisateur pour follow/unfollow:", userID);
    
      const user = users.find((u) => u.id === userID);
    
      if (user?.isFollowing) {
        console.log("❌ Tentative de désabonnement de:", userID);
        await unfollowUser(userID);
      } else {
        console.log("✅ Tentative de suivi de:", userID);
        await followUser(userID);
      }
    
      setUsers((prevUsers) =>
        prevUsers.map((u) =>
          u.id === userID ? { ...u, isFollowing: !u.isFollowing } : u
        )
      );
    
      return true; 
    } catch (error) {
      console.error("❌ Erreur lors du suivi/désabonnement:", error);
      return false; 
    }
  };
  
  
  const handleSearch = async (query) => {
    if (!query) {
      setFilteredUsers(users); 
      return;
    }
    try {
      const results = await apiRequest(`/search_users?query=${query}`);
      setFilteredUsers(results);
    } catch (error) {
      console.error("Error searching users:", error);
    }
  };

  const handleNextPage = () => {
    setPagination((prev) => ({ ...prev, page: prev.page + 1 }));
    fetchUsers();
  };

  const handlePrevPage = () => {
    if (pagination.page > 1) {
      setPagination((prev) => ({ ...prev, page: prev.page - 1 }));
      fetchUsers();
    }
  };

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

  const handleEditProfile = () => {
    router.push("/profil/updateProfil");
  };


  return (
    <div className="min-h-screen bg-gray-900 text-gray-200">
      {/* Header */}
      <header className="w-full px-4 py-3 bg-gray-800 shadow-md flex items-center justify-between">
        <motion.button
          onClick={() => router.push("/")}
          whileHover={{ scale: 1.05 }}
          className="flex items-center text-cyan-400 hover:text-cyan-300"
        >
          <IconArrowLeft className="h-5 w-5 mr-2" />
          Retour à l'accueil
        </motion.button>
        <div className="p-4 bg-gray-700 rounded-lg shadow-md">
              <h3 className="text-xl font-semibold text-cyan-300">Demandes de suivi :</h3>
              {loadingRequests ? (
                <p className="text-gray-300">Chargement...</p>
              ) : followRequests.length > 0 ? (
                <div className="space-y-4">
                  {followRequests.map((req) => (
                    <div key={req.id} className="flex items-center justify-between bg-gray-800 p-4 rounded-lg shadow-md">
                      <div className="flex items-center space-x-4">
                        <img src={req.avatar || "/avatar.png"} alt={req.username} className="w-12 h-12 rounded-full" />
                        <h4 className="text-lg font-semibold text-cyan-300">{req.username}</h4>
                      </div>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => handleAcceptFollow(req.id)}
                          className="bg-green-500 text-white px-4 py-2 rounded-lg hover:bg-green-400"
                        >
                          Accepter
                        </button>
                        <button
                          onClick={() => handleDeclineFollow(req.id)}
                          className="bg-red-500 text-white px-4 py-2 rounded-lg hover:bg-red-400"
                        >
                          Refuser
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-300">Aucune demande en attente.</p>
              )}
            </div>
        <button
          onClick={() => {
            setShowUsersList(!showUsersList);
            if (!showUsersList) fetchUsers();
          }}
          className="bg-cyan-600 text-white px-4 py-2 rounded-lg hover:bg-cyan-500"
        >
          {showUsersList ? "Retour au Profil" : "Liste des Utilisateurs"}
        </button>
      </header>

      {/* Main */}
      <main className="max-w-4xl mx-auto my-8 px-4 sm:px-8 py-6 bg-gray-800 rounded-xl">
        {showUsersList ? (
          <section>
            <SearchBar onSearch={handleSearch} />
            <h1 className="text-2xl font-bold text-cyan-400 mb-4">Liste des utilisateurs</h1>
            {usersLoading ? (
              <p>Chargement...</p>
            ) : (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                  {filteredUsers.map((user) => (
                    <UserCard
                    key={user.id}
                    user={user}
                    onToggleFollow={toggleFollow}  
                  />
                  ))}
                </div>
                <div className="flex justify-between mt-4">
                  <button
                    onClick={handlePrevPage}
                    disabled={pagination.page === 1}
                    className="bg-gray-700 text-white px-4 py-2 rounded-lg"
                  >
                    Précédent
                  </button>
                  <button
                    onClick={handleNextPage}
                    className="bg-cyan-600 text-white px-4 py-2 rounded-lg"
                  >
                    Suivant
                  </button>
                </div>
              </>
            )}
          </section>  
        ) : (

          <section>

            {/* Profil Header */}
            <div className="flex items-center space-x-4">
              <motion.img
                src={profile?.image_profil?.String || "/avatar.png"}
                alt="Avatar"
                className="w-24 h-24 sm:w-28 sm:h-28 rounded-full border-4 border-cyan-500 object-cover shadow-lg"
                whileHover={{ scale: 1.05 }}
              />
              <h2 className="text-3xl font-bold text-cyan-400">Mon Profil</h2>
            </div>

            {/* Profil Info */}
            <div className="mt-8 space-y-6">
              <motion.button
                onClick={handleEditProfile}
                whileHover={{ scale: 1.05 }}
                className="flex items-center space-x-2 bg-cyan-600 text-white px-4 py-2 rounded-lg shadow-md hover:bg-cyan-500 transition-all duration-300"
              >
                <IconEdit className="h-5 w-5" />
                <span className="text-sm font-semibold">Éditer</span>
              </motion.button>

              {[
                { label: "Nom d'utilisateur", value: profile?.username },
                { label: "Nom complet", value: `${profile?.first_name?.String || ""} ${profile?.last_name?.String || ""}` },
                { label: "Biographie", value: profile?.bio || "Pas de biographie" },
                { label: "Confidentialité", value: profile?.is_private ? "Privé" : "Public" },
              ].map((item, idx) => (
                <div key={idx} className="p-4 bg-gray-700 rounded-lg shadow-md">
                  <h3 className="text-xl font-semibold text-cyan-300">{item.label} :</h3>
                  <p className="text-lg text-gray-300">{item.value}</p>
                </div>
              ))}

              <div className="p-4 bg-gray-700 rounded-lg shadow-md">
                <h3 className="text-xl font-semibold text-cyan-300">Abonnés :</h3>
                <p className="text-lg text-gray-300">{profile?.followers_count || 0} abonnés</p>  
              </div>

              <div className="p-4 bg-gray-700 rounded-lg shadow-md">
                <h3 className="text-xl font-semibold text-cyan-300">Abonnements :</h3>
                <p className="text-lg text-gray-300">{profile?.following_count || 0} abonnements</p>
              </div>


              <div className="p-4 bg-gray-700 rounded-lg shadow-md">
                <h3 className="text-xl font-semibold text-cyan-300">Publications :</h3>
                {profile?.posts?.length > 0 ? (
                  <div className="space-y-4">
                    {profile.posts.map((post) => (
                      <div key={post.id} className="bg-gray-800 p-4 rounded-lg shadow-md">
                        <h4 className="text-lg font-semibold text-cyan-300">{post.title}</h4>
                        <p className="text-gray-300">{post.content}</p>
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

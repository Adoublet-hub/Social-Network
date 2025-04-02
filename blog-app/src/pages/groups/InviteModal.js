import { useState, useEffect } from "react";

function InviteModal({ groupId, onClose }) {
  const [users, setUsers] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedUser, setSelectedUser] = useState(null);
  const [isSearching, setIsSearching] = useState(false);

  useEffect(() => {
    if (searchQuery.trim().length === 0) {
      setUsers([]);
      return;
    }

    const fetchUsers = async () => {
      setIsSearching(true);
      try {
        const url = `http://127.0.0.1:8079/users?query=${searchQuery}`;
        const response = await fetch(url, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("authToken")}`,
          },
        });
        if (!response.ok) {
          const errorData = await response.json();
          console.error("Erreur backend :", errorData.error);
          throw new Error("Erreur lors de la recherche d'utilisateurs");
        }
        const data = await response.json();
        setUsers(data);
      } catch (error) {
        console.error("Erreur lors de la recherche :", error);
      } finally {
        setIsSearching(false);
      }
    };
    
    
    

    const delayDebounceFn = setTimeout(() => {
      fetchUsers();
    }, 300); 

    return () => clearTimeout(delayDebounceFn); 
  }, [searchQuery]);

  const handleInvite = async () => {
    if (!selectedUser) return;
  
    console.log("Envoi de l'invitation avec :", {
      groupId,
      invitee_id: selectedUser.id
    });
  
    try {
      const response = await fetch(`http://127.0.0.1:8079/groups/${groupId}/invit_group`, {
        method: "POST",
        headers: {
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          group_id: groupId, 
          receiver_id: selectedUser.id 
        }),
      });
  
      if (!response.ok) throw new Error("Erreur lor s de l'invitation");
  
      alert("Invitation envoyée avec succès");
      onClose();
    } catch (error) {
      console.error("Erreur lors de l'invitation :", error);
      alert("Échec de l'envoi de l'invitation");
    }
  };
  

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div className="bg-gray-800 text-gray-300 p-6 rounded-lg shadow-lg w-96">
        <h2 className="text-lg font-bold text-white mb-4">Inviter un utilisateur</h2>
        <input
          type="text"
          placeholder="Rechercher un utilisateur"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full mb-4 p-2 bg-gray-700 rounded-lg text-gray-300 border border-gray-600"
        />
        {isSearching ? (
          <p className="text-gray-400">Recherche en cours...</p>
        ) : (
          <ul className="mt-4 space-y-2">
            {users.map((user) => (
              <li
                key={user.id}
                className="flex items-center justify-between p-2 bg-gray-700 rounded-lg"
              >
                <div className="flex items-center space-x-3">
                  <img
                    src={user.avatar || "/avatar.png"}
                    alt={user.username || "Utilisateur"}
                    className="h-10 w-10 rounded-full"
                  />
                  <span>{user.username}</span>
                </div>
                <button
                  onClick={() => setSelectedUser(user)}
                  className={`px-4 py-2 rounded-lg ${
                    selectedUser?.id === user.id
                      ? "bg-green-500 text-white"
                      : "bg-gray-600 hover:bg-gray-500"
                  }`}
                >
                  {selectedUser?.id === user.id ? "Sélectionné" : "Sélectionner"}
                </button>
              </li>
            ))}
          </ul>
        )}
        <div className="mt-4 flex justify-end space-x-2">
          <button
            onClick={onClose}
            className="bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-500"
          >
            Annuler
          </button>
          <button
            onClick={handleInvite}
            disabled={!selectedUser}
            className={`px-4 py-2 rounded-lg ${
              selectedUser
                ? "bg-cyan-600 text-white hover:bg-cyan-500"
                : "bg-gray-600 text-gray-400 cursor-not-allowed"
            }`}
          >
            Inviter
          </button>
        </div>
      </div>
    </div>
  );
}

export default InviteModal;

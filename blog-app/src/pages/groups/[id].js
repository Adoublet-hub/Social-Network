import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/router";
import InviteModal from "./InviteModal";
import MessageInputBar from "./MessageInputBar";
import { IconCalendarEvent } from "@tabler/icons-react";
import { IconArrowLeft } from "@tabler/icons-react";
import { motion } from "framer-motion";
import { connectWebSocket, sendMessage, closeWebSocket } from "./websocketService";
import EventManager from "./EventManager";
import GroupPosts from "./GroupPosts"
import PostModal from "./PostModal";





export default function GroupChatPage() {
  const router = useRouter();
  const { id } = router.query;
  const [groupDetails, setGroupDetails] = useState(null);
  const [members, setMembers] = useState([]);
  const [messages, setMessages] = useState([]);
  const [eventDetails, setEventDetails] = useState({ 
    title: "", 
    description: "", 
    dateTime: "", 
    group_id: "" 
  });
  
  const [isCreatingEvent, setIsCreatingEvent] = useState(false);
  const [loading, setLoading] = useState(true);
  const [clientID, setClientID] = useState("");
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const messagesEndRef = useRef(null);
  const [isPostModalOpen, setIsPostModalOpen] = useState(false);




/*-------------------------------------------------------------------------------------------------------------------------- */

  useEffect(() => {
    if (id) {
      setEventDetails((prev) => ({ ...prev, group_id: id }));
    }
  }, [id]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);
  
  useEffect(() => {
    const storedUserID = localStorage.getItem("authToken")
    if (storedUserID) {
      setClientID(storedUserID);
    } else {
      console.warn("⚠️ Aucun ID utilisateur trouvé.");
    }

    const token = localStorage.getItem("authToken");
    if (token) {
      connectWebSocket(
        token,
        (data) => handleWebSocketMessage(data),
        () => console.log("WebSocket ouvert"),
        () => console.log("WebSocket fermé")
      );
    }

    return () => {
      closeWebSocket();
    };
  }, [id]);

  const handleWebSocketMessage = (data) => {
    if (!data.type || !data.sender_username) {
      console.warn("⚠️ Message invalide reçu :", data);
      return;
    }

    switch (data.type) {
      case "newMessage":
        setMessages((prev) => {
          if (!prev.some((msg) => msg.id === data.id)) {
            return [...prev, data];
          }
          return prev;
        });
        break;
      case "newImage":
        setMessages((prev) => [...prev, data]);
        break;
      default:
        console.warn("⚠️ Type de message inconnu :", data.type);
    }
  };

  const handleSendMessage = async (message) => {
    if (!message.trim()) {
      console.warn("⚠️ Message vide non autorisé.");
      return;
    }
  
    const payload = {
      content: message,
      target_username: id,
    };
  
    try {
      const response = await fetchWithAuth(`http://127.0.0.1:8079/messagegroup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
  
      if (!response.ok) {
        console.error("Erreur lors de l'envoi du message.");
        return;
      }
  
      const data = await response.json().catch(() => ({})); 
      if (data.status === "success") {
        console.log("Message envoyé !");
      } else {
        console.error("Erreur lors de l'envoi du message :", data);
      }
    } catch (err) {
      console.error("Erreur lors de l'envoi du message :", err);
    }
  };

  const fetchWithAuth = async (url, options = {}) => {
    const token = localStorage.getItem("authToken");
    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        Authorization: `Bearer ${token}`,
      },
    });
  };

  const fetchMessagesGroup = async (username, offset = 0) => {
    try {
      const response = await fetchWithAuth(`http://127.0.0.1:8079/messagegroup?user=${id}&offset=0`);
      if (!response.ok) throw new Error("Erreur lors de la récupération des messages.");
  
      const data = await response.json();
  
      if (Array.isArray(data)) {
        setMessages((prevMessages) => [...data, ...prevMessages]);
      } else {
        console.warn("La réponse des messages n'est pas un tableau :", data);
      }
    } catch (err) {
      console.error("Erreur lors de la récupération des messages :", err);
    }
  };

  useEffect(() => {
    if (id) {
      fetchMessagesGroup();
    }
  }, [id]);

  const handleSendImage = (imageFile) => {
    if (!imageFile) {
      console.warn("Aucune image sélectionnée.");
      return;
    }
  
    const reader = new FileReader();
    reader.onload = () => {
      const imageBase64 = reader.result.split(",")[1]; // Convertir en base64
  
      const message = {
        type: "newImage",
        content: imageBase64,
        group_id: id,
        sender_id: clientID,
        fileName: imageFile.name,
        fileType: imageFile.type,
      };
  
      sendMessage(message); // Envoi via WebSocket
      console.log("📤 Image envoyée via WebSocket :", message);
    };
  
    reader.onerror = () => {
      console.error("Erreur lors de la lecture de l'image.");
    };
  
    reader.readAsDataURL(imageFile);
  };
  
  
/*------------------------------------------------------------------------------------------------------------------------ */

  useEffect(() => {
    if (!id) return;
  
    const fetchGroupData = async () => {
      try {
        setLoading(true);
        const response = await fetch(`http://127.0.0.1:8079/group/${id}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("authToken")}`,
          },
          credentials: "include",
        });
    
        if (!response.ok) {
          let errorMessage = "Erreur lors du chargement du groupe.";
          try {
            const errorData = await response.json();
            errorMessage = errorData.error || errorMessage;
          } catch (err) {
            console.warn("Réponse vide ou invalide reçue de l'API.");
          }
          throw new Error(errorMessage);
        }
    
        const data = await response.json().catch(() => ({ messages: [], group: {}, members: [] }));
    
        setMessages((prevMessages) => [
          ...prevMessages,
          ...(Array.isArray(data.messages) ? data.messages : []),
        ]);
        setGroupDetails(data.group || {});
        setMembers(data.members || []);
      } catch (error) {
        console.error("Erreur de chargement :", error);
        alert("Impossible de charger le groupe. Veuillez réessayer.");
      } finally {
        setLoading(false);
      }
    };
    
  
    fetchGroupData();
  }, [id]);

  return (
    <div className="flex h-screen bg-gray-900 text-gray-300">
      {/* Sidebar */}
      <div className="w-full lg:w-1/3 bg-gray-800 flex flex-col border-r border-gray-700">
        <header className="p-4 flex items-center justify-between bg-gray-800 shadow-md">
        <motion.button
            onClick={() => router.push("/")}
            whileHover={{ scale: 1.05 }}
            className="text-cyan-400 hover:text-cyan-300 duration-200"
          >
            <IconArrowLeft className="h-5 w-5 mr-2" />
            Retour à l'accueil
          </motion.button>
          <h1 className="text-lg font-bold text-white truncate">
            {groupDetails?.name || "Chargement..."}
          </h1>
          <button
            onClick={() => setIsCreatingEvent(!isCreatingEvent)}
            className="text-cyan-400 hover:text-cyan-300"
          >
            <IconCalendarEvent />
          </button>
          <button
            onClick={() => setIsPostModalOpen(!isPostModalOpen)}
            className="text-cyan-400 hover:text-cyan-300"
          >
            📋
          </button>
        </header>

        <div className="flex-1 overflow-y-auto p-4">
          <h2 className="text-sm font-semibold text-gray-400 mb-2">Membres du groupe</h2>
          {members.length > 0 ? (
            members.map((member) => (
              <div
                key={member.user_id}
                className="flex items-center justify-between p-2 bg-gray-700 rounded-lg mb-2 shadow-sm"
              >
                <div className="flex items-center space-x-3">
                  <img
                    src={member.image_profil?.String || "/avatar.png"}
                    alt={member.username || "Utilisateur"}
                    className="h-10 w-10 rounded-full shadow-md"
                  />
                  <p className="text-sm font-medium text-white truncate">
                    {member.user_id === clientID ? `Moi (${member.username})` : member.username}
                  </p>
                </div>
              </div>
            ))
          ) : (
            <p className="text-gray-500">Aucun membre trouvé.</p>
          )}
        </div>
        <button
          onClick={() => setShowInviteModal(true)}
          className="w-full bg-cyan-600 text-white py-2 rounded-lg hover:bg-cyan-500 mt-4"
        >
          Inviter des utilisateurs
        </button>
      </div>

      {/* Messages and Main Content */}
      <div className="flex-1 flex flex-col">
        <header className="p-4 bg-gray-800 shadow-md flex justify-between items-center">
          <h1 className="text-lg font-bold text-white truncate">
            {groupDetails?.name || "Discussion de groupe"}
          </h1>
        </header>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 bg-gray-900 space-y-4">

        {messages?.length > 0 ? (
          messages.map((message, index) => (
            <motion.div
              key={`${message.id}-${index}`}
              className={`flex ${
                message.sender_username === clientID ? "justify-end" : "justify-start"
              }`}
            >
              {message.sender_username !== clientID && (
                <img
                  src={message.avatar || "/avatar.png"}
                  alt={message.sender_username}
                  className="w-8 h-8 rounded-full shadow-md mr-2"
                />
              )}
              
              <div
                className={`p-3 max-w-xs rounded-2xl shadow-md text-sm ${
                  message.sender_username === clientID
                    ? "bg-cyan-600 text-white self-end rounded-br-none"
                    : "bg-gray-700 text-gray-200 self-start rounded-bl-none"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-bold text-xs">
                    {message.sender_username === clientID ? "Moi" : message.sender_username}
                  </span>
                  <span className="text-xs text-gray-400 ml-2">
                    {new Date(message.timestamp).toLocaleTimeString("fr-FR", {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </span>
                </div>

                {/* Condition pour afficher l'image si c'est un message d'image */}
                {message.type === "newImage" ? (
                  <img
                    src={`data:${message.fileType};base64,${message.content}`}
                    alt={message.fileName || "Image"}
                    className="mt-2 rounded-lg max-w-full max-h-60"
                  />
                ) : (
                  <p className="break-words">{message.content}</p>
                )}
              </div>

              {message.sender_username === clientID && (
                <img
                  src={message.avatar || "/avatar.png"}
                  alt={message.sender_username}
                  className="w-8 h-8 rounded-full shadow-md ml-2"
                />
              )}
            </motion.div>
          ))
        ) : (
          <p className="text-gray-500 text-center">Aucun message pour l'instant.</p>
        )}


          <div ref={messagesEndRef} />

          {isCreatingEvent && (
            <EventManager groupId={id} />
          )}
          {isPostModalOpen && (
            <PostModal groupId={id} onClose={() => setIsPostModalOpen(false)} />
          )}
        </div>

        {/* Barre d'entrée */}
        <MessageInputBar onSendMessage={handleSendMessage} onSendImage={handleSendImage} />

        {/* Modale d'invitation */}
        {showInviteModal && (
          <InviteModal
            groupId={id}
            onClose={() => setShowInviteModal(false)}
          />
        )}
      </div>
    </div>
  );
}  
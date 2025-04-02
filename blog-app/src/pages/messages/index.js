import { useState, useEffect, useRef } from "react";
import { motion } from "framer-motion";
import { IconArrowLeft, IconCircleCheck, IconMessageCircle, IconUser } from "@tabler/icons-react";
import MessageInputBar from "../groups/MessageInputBar";

export default function MessagesPage() {
  const [clientID, setClientID] = useState("");
  const [users, setUsers] = useState([]);
  const [onlineUsers, setOnlineUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [isTyping, setIsTyping] = useState(false);
  const [usersLoading, setUsersLoading] = useState(false);

  const socketRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 10;

  useEffect(() => {
    connectWebSocket();

    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, [selectedUser]);

  const connectWebSocket = () => {
    // Si une connexion est déjà ouverte, ne rien faire
    if (socketRef.current && (socketRef.current.readyState === WebSocket.OPEN || socketRef.current.readyState === WebSocket.CONNECTING)) {
      console.log("✅ WebSocket déjà connecté ou en cours de connexion.");
      return;
    }
  
    const token = localStorage.getItem("authToken");
    socketRef.current = new WebSocket(`ws://localhost:8079/ws?token=${token}`);
  
    socketRef.current.onopen = () => {
      console.log("✅ WebSocket connecté");
      reconnectAttemptsRef.current = 0;
    };
  
    socketRef.current.onmessage = (event) => {
      if (!event.data || event.data.trim() === "") {
        console.warn("⚠️ Message vide reçu via WebSocket.");
        return;
      }
    
      try {
        const data = JSON.parse(event.data);
        console.log("📩 Nouveau message reçu via WebSocket :", data);
        handleWebSocketMessage(data);
      } catch (err) {
        console.error("❌ Erreur lors du parsing du message WebSocket :", err);
      }
    };
    
    
    socketRef.current.onerror = (error) => {
      console.error("❌ Erreur WebSocket :", error);
      socketRef.current.close();
    };
  };
  
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
    if (!selectedUser || !message.trim()) {
      console.warn("⚠️ Utilisateur ou message manquant.");
      return;
    }
  
    try {
      const payload = {
        sender_username: clientID,
        target_username: selectedUser.username,
        content: message,
      };
  
      const response = await fetchWithAuth("http://127.0.0.1:8079/message", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        console.error("❌ Erreur lors de l'envoi du message.");
        return;
      }
  
      // ✅ Vérifie que la réponse contient un body valide
      const data = await response.json().catch(() => ({})); // ✅ Ajoute un catch pour éviter l'erreur
      if (data.status === "success") {
        console.log("✅ Message envoyé !");
      } else {
        console.error("❌ Erreur lors de l'envoi du message :", data);
      }
    } catch (err) {
      console.error("Erreur lors de l'envoi du message :", err);
    }
  };

  const formatTimestamp = (timestamp) => {
    if (!timestamp) return "Heure inconnue";
  
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return "Heure invalide";
  
    return date.toLocaleString("fr-FR", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
    });
  };  

  
  const handleSendImage = (image) => {
    const reader = new FileReader();
    reader.readAsDataURL(image);
    reader.onload = () => {
      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        const payload = {
          type: "newImage",
          senderUsername: clientID,
          targetUsername: selectedUser?.username,
          content: reader.result,
        };
        socketRef.current.send(JSON.stringify(payload));
      } else {
        console.warn("⚠️ WebSocket non connecté. Tentative de reconnexion...");
        connectWebSocket();
      }
    };
  };

  useEffect(() => {
    fetchUsers();
    fetchOnlineUsers();
  }, []);

  const fetchMessages = async (username, offset = 0) => {
    try {
      const response = await fetchWithAuth(`http://127.0.0.1:8079/message?user=${username}&offset=${offset}`);
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
  

  const handleSelectUser = (user) => {
    setSelectedUser(user);
    setMessages([]);
    fetchMessages(user.username);
  };

  const fetchUsers = async () => {
    setUsersLoading(true);
    try {
      const response = await fetchWithAuth("http://127.0.0.1:8079/list_amis?page=1&limit=10");
      const data = await response.json();
      setUsers(data.results || []);
    } catch (err) {
      console.error("Erreur lors de la récupération des utilisateurs :", err);
    } finally {
      setUsersLoading(false);
    }
  };

  const fetchOnlineUsers = async () => {
    try {
      const response = await fetchWithAuth("http://127.0.0.1:8079/online");
      const data = await response.json();
      setOnlineUsers(data);
    } catch (err) {
      console.error("Erreur lors de la récupération des utilisateurs en ligne :", err);
    }
  };
  /*
  const handleTyping = () => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ type: "typing", username: selectedUser?.username }));
    }
  };
  */

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

  return (
    <div className="flex h-screen bg-gray-900 text-gray-300">
      <div className="w-full lg:w-1/3 bg-gray-800 flex flex-col border-r border-gray-700">
        <header className="p-4 flex items-center justify-between bg-gray-800 shadow-md">
          <motion.button
            onClick={() => (window.location.href = "/")}
            whileHover={{ scale: 1.05 }}
            className="text-cyan-400 hover:text-cyan-300 flex items-center"
          >
            <IconArrowLeft className="h-5 w-5 mr-2" />
            Accueil
          </motion.button>
        </header>

        <div className="flex-1 overflow-y-auto p-4">
          <h2 className="text-sm font-semibold text-gray-400 mb-2">Utilisateurs</h2>
          {usersLoading ? (
            <p>Chargement des utilisateurs...</p>
          ) : (
          <ul>
            {users.map((user, index) => (
              <motion.li
                key={user.id || `${user.username}-${index}`} 
                onClick={() => handleSelectUser(user)}
                whileHover={{ scale: 1.02 }}
                className={`p-2 bg-gray-700 rounded-lg mb-2 cursor-pointer ${
                  selectedUser?.username === user.username ? "bg-cyan-600 text-white" : ""
                }`}
              >
                <IconUser className="inline-block mr-2 text-gray-400" />
                {user.username} {onlineUsers.includes(user.username) ? <IconCircleCheck className="inline-block text-green-500 ml-2" /> : "⚪"}
              </motion.li>
            ))}
          </ul>
          )}
        </div>
      </div>

      <div className="flex-1 flex flex-col">
        {/* Header de la conversation */}
        <header className="p-4 bg-gray-800 shadow-md flex justify-between items-center">
          <h1 className="text-lg font-bold text-white truncate">
            {selectedUser ? `Discussion avec ${selectedUser.username}` : "Aucune discussion"}
          </h1>
          {isTyping && <IconMessageCircle className="text-gray-400 animate-pulse" />}
        </header>

        {/* Zone des messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-900">
        {messages.length > 0 ? (
          [...new Map(messages.map((msg) => [msg.id, msg])).values()].map((message, index) => (
            <motion.div
              key={`${message.id || "msg"}-${index}`} 
              className={`p-3 rounded-3xl max-w-xs shadow-md ${
                message.sender_username === clientID ? "bg-cyan-600 text-white self-end" : "bg-gray-700 self-start"
              }`}
            >
              {message.sender_username !== clientID && (
                <p className="font-medium text-sm text-cyan-400 mb-1">{message.sender_username}</p>
              )}
              
              <p className="break-words text-sm">{message.content}</p>

              <small className="block mt-1 text-xs text-gray-400 text-right">
                {formatTimestamp(message.timestamp)}
              </small>
            </motion.div>
          ))
        ) : (
          <p className="text-gray-500 text-center">Aucun message pour l'instant.</p>
        )}
        </div>


        <MessageInputBar onSendMessage={handleSendMessage} onSendImage={handleSendImage} />
      </div>
    </div>
  );
}

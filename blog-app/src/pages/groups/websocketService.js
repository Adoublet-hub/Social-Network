let socket = null;
let reconnectAttempts = 0;

export const connectWebSocket = (token, onMessageCallback, onOpenCallback, onCloseCallback) => {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    console.log("✅ WebSocket déjà connecté ou en cours de connexion.");
    return;
  }

  socket = new WebSocket(`ws://localhost:8079/ws?token=${token}`);

  socket.onopen = () => {
    console.log("✅ WebSocket connecté");
    reconnectAttempts = 0;
    if (onOpenCallback) onOpenCallback();
  };

  socket.onmessage = (event) => {
    if (!event.data || event.data.trim() === "") {
      console.warn("⚠️ Message vide reçu via WebSocket.");
      return;
    }

    try {
      const data = JSON.parse(event.data);
      console.log("📩 Nouveau message reçu via WebSocket :", data);
      if (onMessageCallback) onMessageCallback(data);
    } catch (err) {
      console.error("❌ Erreur lors du parsing du message WebSocket :", err);
    }
  };

  socket.onerror = (error) => {
    console.error("❌ Erreur WebSocket :", error);
    socket.close();
  };

  socket.onclose = () => {
    console.warn("⚠️ WebSocket déconnecté, tentative de reconnexion...");
    if (reconnectAttempts < 5) {
      setTimeout(() => {
        reconnectAttempts++;
        connectWebSocket(token, onMessageCallback, onOpenCallback, onCloseCallback);
      }, 3000);
    }
    if (onCloseCallback) onCloseCallback();
  };
};

export const sendMessage = (message) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(message));
  } else {
    console.warn("⚠️ WebSocket non connecté. Impossible d'envoyer le message.");
  }
};

export const closeWebSocket = () => {
  if (socket) {
    socket.close();
    socket = null;
  }
};

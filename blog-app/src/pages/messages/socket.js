import { Server } from "ws";

let wss;

export default function handler(req, res) {
  if (!res.socket.server.wss) {
    console.log("✅ Initialisation du WebSocket Server...");

    wss = new Server({ noServer: true });
    res.socket.server.wss = wss;

    wss.on("connection", (socket) => {
      console.log("✅ Nouvel utilisateur connecté");
    
      socket.on("message", (message) => {
        try {
          const data = JSON.parse(message);
          console.log("📥 Message reçu :", data);
    
          if (!["typing", "newMessage", "newImage"].includes(data.type)) {
            console.warn("⚠️ Type de message non valide :", data.type);
            return;
          }
    
          wss.clients.forEach((client) => {
            if (client.readyState === 1) {
              client.send(JSON.stringify(data));
            }
          });
        } catch (error) {
          console.error("❌ Erreur lors du traitement du message :", error);
        }
      });
    
      socket.on("close", () => {
        console.log("⚠️ Utilisateur déconnecté");
      });
    });    
      
  }

  res.end();
}

export const config = {
  api: {
    bodyParser: false,
  },
};

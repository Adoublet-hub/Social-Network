import { useEffect, useState } from "react";
import { IconCalendarEvent, IconClock, IconCheck, IconX } from "@tabler/icons-react";

export default function EventList({ groupId }) {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [userVotes, setUserVotes] = useState({});

  const fetchEvents = async () => {
    try {
      const response = await fetch(`http://127.0.0.1:8079/list_event?group_id=${groupId}`, {
        headers: {
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
      });

      if (!response.ok) {
        throw new Error("Erreur lors de la récupération des événements.");
      }

      const data = await response.json();
      setEvents(data);
    } catch (error) {
      console.error("Erreur de récupération des événements :", error);
      setError("Impossible de charger les événements.");
    } finally {
      setLoading(false);
    }
  };

  const fetchUserVotes = async () => {
    try {
      const response = await fetch(`http://127.0.0.1:8079/get_user_votes?group_id=${groupId}`, {
        headers: {
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
      });

      if (!response.ok) throw new Error("Erreur lors de la récupération des votes.");

      const data = await response.json();

      if (!Array.isArray(data)) {
        console.error("Données de votes invalides :", data);
        return;
      }

      const votesMap = {};
      data.forEach((vote) => {
        votesMap[vote.event_id] = vote.response;
      });

      setUserVotes(votesMap);
    } catch (error) {
      console.error("Erreur de récupération des votes :", error);
    }
  };

  useEffect(() => {
    if (!groupId) return;
    fetchEvents();
    fetchUserVotes();
  }, [groupId]);

  const respondToEvent = async (eventId, response) => {
    try {

      const formattedResponse = response === "Going" ? "Going" : "Not going";
  
      const res = await fetch(`http://127.0.0.1:8079/respond_to_event`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: JSON.stringify({ event_id: eventId, response: formattedResponse }),
      });
  
      if (!res.ok) {
        const errorText = await res.text();
        console.error("Server Error Response:", errorText);
        throw new Error("Erreur lors de la réponse à l'événement.");
      }
  
      const updatedEvent = await res.json();
  
      setUserVotes((prevVotes) => ({
        ...prevVotes,
        [eventId]: formattedResponse,
      }));
  
      setEvents((prevEvents) =>
        prevEvents.map((event) => {
          if (event.id === eventId) {
            let newGoing = event.options?.Going || 0;
            let newNotGoing = event.options?.NotGoing || 0;
      
            if (userVotes[eventId] === formattedResponse) {
              formattedResponse === "Going" ? newGoing-- : newNotGoing--;
            } else if (userVotes[eventId]) {
              formattedResponse === "Going" ? (newGoing++, newNotGoing--) : (newNotGoing++, newGoing--);
            } else {
              formattedResponse === "Going" ? newGoing++ : newNotGoing++;
            }
      
            return {
              ...event,
              options: { Going: newGoing, NotGoing: newNotGoing },
            };
          }
          return event;
        })
      );
      
    } catch (error) {
      console.error("❌ Erreur lors de l'envoi de la réponse :", error);
      alert(error.message);
    }
  };
  
  
  
  
  if (loading) return <p className="text-gray-500 text-center">Chargement des événements...</p>;
  if (error) return <p className="text-red-500 text-center">{error}</p>;

  return (
    <div className="space-y-3">
      {events?.length > 0 ? (
        events.map((event) => (
          <div key={event.id} className="p-4 bg-gray-800 border-l-4 border-cyan-400 rounded-lg shadow-md max-w-sm mx-auto">
            <div className="flex justify-between">
              <h3 className="text-base font-semibold text-white">
                <IconCalendarEvent className="inline-block text-cyan-400 mr-2" size={18} /> {event.title}
              </h3>
              <p className="text-xs text-gray-400 flex items-center">
                <IconClock className="mr-1" size={14} />
                {event.event_date ? new Date(event.event_date).toLocaleString() : "Date non définie"}
              </p>
            </div>

            <p className="text-sm text-gray-300 mt-1">{event.description}</p>

            {/* Affichage des votes */}
            <div className="mt-2 flex justify-between text-xs text-gray-400">
              <span className="text-green-400">Going: {event.options?.Going}</span>
              <span className="text-red-400">Not Going: {event.options?.NotGoing}</span>
            </div>

            {/* Boutons pour voter */}
            <div className="mt-2 flex justify-between">
            <button
              onClick={() => respondToEvent(event.id, "Going")}
              className={`px-3 py-1 text-xs font-semibold rounded transition-transform transform ${
                userVotes[event.id] === "Going"
                  ? "bg-green-700 text-white scale-105"
                  : "bg-green-600 text-white hover:bg-green-500"
              }`}
            >
              <IconCheck size={14} className="mr-1" />
              Going
            </button>

              <button
                onClick={() => respondToEvent(event.id, "NotGoing")}
                className={`px-3 py-1 text-xs font-semibold rounded flex items-center ${
                  userVotes[event.id] === "NotGoing" ? "bg-red-700 text-white" : "bg-red-600 text-white hover:bg-red-500"
                }`}
              >
                <IconX size={14} className="mr-1" />
                Not Going
              </button>
            </div>
          </div>
        ))
      ) : (
        <p className="text-gray-500 text-center">Aucun événement prévu.</p>
      )}
    </div>
  );
}

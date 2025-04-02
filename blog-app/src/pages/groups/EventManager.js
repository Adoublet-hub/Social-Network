import { useState } from "react";
import EventList from "./EventList";
import { IconPlus, IconX } from "@tabler/icons-react";

export default function EventManager({ groupId }) {
  const [isCreatingEvent, setIsCreatingEvent] = useState(false);
  const [eventDetails, setEventDetails] = useState({
    title: "",
    description: "",
    dateTime: "",
  });

  const handleCreateEvent = async () => {
    try {
      if (!eventDetails.title.trim() || !eventDetails.description.trim() || !eventDetails.dateTime) {
        alert("Tous les champs doivent être remplis.");
        return;
      }

      const formattedDate = new Date(eventDetails.dateTime).toISOString();

      const response = await fetch(`http://127.0.0.1:8079/group/${groupId}/create_event`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: JSON.stringify({
          title: eventDetails.title,
          description: eventDetails.description,
          event_date: formattedDate,
          group_id: groupId,
        }),
      });

      if (!response.ok) throw new Error("Erreur lors de la création de l'événement.");

      setIsCreatingEvent(false);
      setEventDetails({ title: "", description: "", dateTime: "" });
    } catch (error) {
      console.error("Erreur lors de la création de l'événement :", error);
    }
  };

  return (
    <div className="mt-4">
      <button
        onClick={() => setIsCreatingEvent(!isCreatingEvent)}
        className="flex items-center justify-center w-full bg-cyan-600 text-white py-2 rounded-lg hover:bg-cyan-500 transition duration-200"
      >
        {isCreatingEvent ? <IconX size={20} /> : <IconPlus size={20} />}
        <span className="ml-2">{isCreatingEvent ? "Annuler" : "Créer un événement"}</span>
      </button>

      {isCreatingEvent && (
        <div className="mt-3 p-4 bg-gray-800 rounded-lg shadow-lg max-w-sm mx-auto">
          <h3 className="text-lg font-semibold text-white mb-3">Créer un événement</h3>
          <input
            type="text"
            placeholder="Titre"
            value={eventDetails.title}
            onChange={(e) => setEventDetails((prev) => ({ ...prev, title: e.target.value }))}
            className="w-full p-2 mb-2 bg-gray-700 text-gray-300 rounded-lg"
          />
          <textarea
            placeholder="Description"
            value={eventDetails.description}
            onChange={(e) => setEventDetails((prev) => ({ ...prev, description: e.target.value }))}
            className="w-full p-2 mb-2 bg-gray-700 text-gray-300 rounded-lg"
          />
          <input
            type="datetime-local"
            value={eventDetails.dateTime}
            onChange={(e) => setEventDetails((prev) => ({ ...prev, dateTime: e.target.value }))}
            className="w-full p-2 mb-3 bg-gray-700 text-gray-300 rounded-lg"
          />
          <button onClick={handleCreateEvent} className="w-full bg-cyan-600 text-white py-2 rounded-lg hover:bg-cyan-500">
            Valider
          </button>
        </div>
      )}

      <EventList groupId={groupId} />
    </div>
  );
}

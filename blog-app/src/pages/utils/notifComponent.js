import React, { useState, useMemo, useEffect } from "react";
import { handleNotificationAction, fetchNotifications } from "../profil/follow";

const Notifications = ({ notifications, setNotifications, onViewProfile }) => {
  const [loading, setLoading] = useState({});
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    const interval = setInterval(async () => {
      const newNotifications = await fetchNotifications();
      setNotifications(newNotifications);
    }, 10000);

    return () => clearInterval(interval); 
  }, []);

  const handleAction = async (action, notificationID, groupID = null) => {
    try {
      setLoading((prev) => ({ ...prev, [notificationID]: true }));
      setErrorMessage("");

      await handleNotificationAction(action, notificationID, groupID);

      setNotifications((prev) =>
        prev.map((notif) =>
          notif.id === notificationID ? { ...notif, read: true } : notif
        )
      );
    } catch (error) {
      setErrorMessage("Impossible de traiter la notification. Réessayez.");
      console.error("Erreur lors de l'action:", error);
    } finally {
      setLoading((prev) => ({ ...prev, [notificationID]: false }));
    }
  };

  const unreadNotifications = useMemo(
    () => notifications.filter((notif) => !notif.read),
    [notifications]
  );

  return (
    <div className="absolute top-16 right-4 w-80 bg-gray-800 p-4 rounded-lg shadow-lg z-50 transition-all duration-300">
      <h3 className="text-lg font-bold text-white mb-2">Notifications</h3>

      {/* Affichage des erreurs si nécessaire */}
      {errorMessage && <p className="text-red-500 text-sm">{errorMessage}</p>}

      {unreadNotifications.length > 0 ? (
        <>
          {/* Bouton pour marquer toutes les notifications comme lues */}
          <button
            className="text-xs text-cyan-400 hover:underline mb-2"
            onClick={() =>
              unreadNotifications.forEach((notif) =>
                handleAction("markAsRead", notif.id)
              )
            }
          >
            Marquer tout comme lu
          </button>

          <ul>
            {unreadNotifications.map((notif) => (
              <li
                key={notif.id}
                className="p-3 mb-2 rounded cursor-pointer transition-all duration-200 bg-gray-600 hover:bg-gray-500"
                onClick={() => onViewProfile(notif.sender_id)}
              >
                <div className="flex items-center space-x-4">
                  <img
                    src={notif.sender_profile_picture || "/avatar.png"}
                    alt={notif.sender_name}
                    className="w-10 h-10 rounded-full"
                  />
                  <div>
                    <p className="text-sm text-white font-semibold">
                      {notif.sender_name}
                    </p>
                    <p className="text-sm text-gray-300">{notif.content}</p>
                  </div>
                </div>

                {/* Gestion des demandes d'ami */}
                {notif.type === "follow_request" && (
                  <div className="flex justify-between items-center mt-2">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleAction("accept", notif.id);
                      }}
                      disabled={loading[notif.id]}
                      className={`text-xs px-3 py-1 rounded font-semibold ${
                        loading[notif.id]
                          ? "bg-gray-500 text-white cursor-not-allowed"
                          : "bg-green-500 hover:bg-green-400 text-white"
                      }`}
                    >
                      {loading[notif.id] ? "Accept..." : "Accepter"}
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleAction("refuse", notif.id);
                      }}
                      disabled={loading[notif.id]}
                      className={`text-xs px-3 py-1 rounded font-semibold ${
                        loading[notif.id]
                          ? "bg-gray-500 text-white cursor-not-allowed"
                          : "bg-red-500 hover:bg-red-400 text-white"
                      }`}
                    >
                      {loading[notif.id] ? "Refuse..." : "Refuser"}
                    </button>
                  </div>
                )}

                {/* Gestion des invitations de groupe */}
                {notif.type === "group_invite" && (
                  <div className="flex justify-between items-center mt-2">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleAction(
                          "accept_group_invite",
                          notif.id,
                          notif.group_id
                        );
                      }}
                      disabled={loading[notif.id]}
                      className={`text-xs px-3 py-1 rounded font-semibold ${
                        loading[notif.id]
                          ? "bg-gray-500 text-white cursor-not-allowed"
                          : "bg-green-500 hover:bg-green-400 text-white"
                      }`}
                    >
                      {loading[notif.id] ? "Joining..." : "Rejoindre"}
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleAction("refuse", notif.id);
                      }}
                      disabled={loading[notif.id]}
                      className={`text-xs px-3 py-1 rounded font-semibold ${
                        loading[notif.id]
                          ? "bg-gray-500 text-white cursor-not-allowed"
                          : "bg-red-500 hover:bg-red-400 text-white"
                      }`}
                    >
                      {loading[notif.id] ? "Refusing..." : "Refuser"}
                    </button>
                  </div>
                )}

                {/* Bouton pour marquer comme lu */}
                {!notif.read && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleAction("markAsRead", notif.id);
                    }}
                    className="text-xs text-cyan-400 hover:underline mt-2 block"
                  >
                    Marquer comme lu
                  </button>
                )}
              </li>
            ))}
          </ul>
        </>
      ) : (
        <p className="text-sm text-gray-400">Aucune notification pour le moment.</p>
      )}
    </div>
  );
};

export default Notifications;

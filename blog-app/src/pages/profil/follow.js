const API_BASE_URL = "http://127.0.0.1:8079";

const apiRequest = async (endpoint, method = "GET", body = null) => {
  const headers = {
    Authorization: `Bearer ${localStorage.getItem("authToken")}`,
  };

  if (!(body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const options = {
    method,
    headers,
    ...(body && { body: body instanceof FormData ? body : JSON.stringify(body) }),
  };

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, options);

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "API request failed");
    }

    return await response.json();
  } catch (error) {
    console.error(`Erreur dans ${endpoint}:`, error);
    throw error;
  }
};

export const followUser = async (userId) => {
  console.log("Envoi d'une requête pour suivre l'utilisateur:", userId);
  return await apiRequest("/follow_request", "POST", { friend_id: userId });
};

export const unfollowUser = async (userId) => {
    console.log("Désabonnement de l'utilisateur:", userId);
    return await apiRequest("/unfollow", "DELETE", { followed_id: userId }); 
};
  

export const acceptFollowRequest = async (requestId) => {
  console.log("Acceptation de la demande de suivi:", requestId);
  return await apiRequest("/accept_follower", "POST", { request_id: requestId });
};

export const declineFollowRequest = async (requestId) => {
  console.log("Refus de la demande de suivi:", requestId);
  return await apiRequest("/decline_follower", "POST", { request_id: requestId });
};

export const handleNotificationAction = async (action, notificationID, groupID = null) => {
  console.log("Action exécutée:", action, "| ID Notification:", notificationID, "| ID Groupe:", groupID || "N/A");

  const actionEndpoints = {
    accept_group_invite: { 
      endpoint: "/accept_group_invite", 
      body: groupID ? { group_id: groupID, notification_id: notificationID } : { notification_id: notificationID }
    },
    markAsRead: { endpoint: "/mark_as_read", body: { notification_id: notificationID } },
    accept: { endpoint: "/accept_follower", body: { request_id: notificationID } },
    refuse: { endpoint: "/decline_follower", body: { request_id: notificationID } },
  };

  const selectedAction = actionEndpoints[action];

  if (!selectedAction) {
    console.error(`⚠️ Action inconnue : ${action}`);
    throw new Error("Action non valide.");
  }

  try {
    const response = await fetch(`${API_BASE_URL}${selectedAction.endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
      body: JSON.stringify(selectedAction.body),
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error(`Erreur API (${response.status}):`, errorText);
      throw new Error(`Erreur API : ${errorText}`);
    }

    console.log(`Succès de l'action ${action} sur la notification ${notificationID}`);
    return await response.json();
  } catch (error) {
    console.error(`Erreur lors de l'exécution de l'action ${action}:`, error);
    throw error;
  }
};

export const fetchNotifications = async () => {
    try {
      console.log("Tentative de récupération des notifications...");
  
      const response = await fetch(`${API_BASE_URL}/notifications`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });
  
      if (!response.ok) {
        const errorText = await response.text();
        console.error(`Erreur API (${response.status}):`, errorText);
        throw new Error(`Erreur API : ${errorText}`);
      }
  
      const data = await response.json();
      console.log("Notifications reçues :", data);
  
      if (!Array.isArray(data)) {
        console.error("⚠️ Format invalide :", data);
        return [];
      }
  
      return data;
    } catch (error) {
      console.error("Erreur récupération notifications:", error);
      return [];
    }
};
  

export const markNotificationAsRead = async (notificationID) => {
  try {
    console.log(`Marquage de la notification ${notificationID} comme lue`);
    return await apiRequest("/mark_as_read", "POST", { notification_id: notificationID });
  } catch (error) {
    console.error(`Erreur lors du marquage de la notification ${notificationID} comme lue:`, error);
    throw error;
  }
};

export const reloadNotifications = async () => {
  try {
    console.log("Rechargement des notifications...");

    const response = await fetch(`${API_BASE_URL}/notifications`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error(`Erreur API (${response.status}):`, errorText);
      throw new Error(`Erreur API : ${errorText}`);
    }

    const data = await response.json();

    if (!Array.isArray(data)) {
      console.error("⚠️ Format des notifications invalide :", data);
      return [];
    }

    console.log("Notifications rechargées avec succès:", data);
    return data;
  } catch (error) {
    console.error("Erreur lors du rechargement des notifications:", error);
    return [];
  }
};

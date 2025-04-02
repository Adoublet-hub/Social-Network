export const fetchUserProfile = async (userId) => {
  try {
    if (!userId) {
      throw new Error("User ID is undefined");
    }
    console.log(`Fetching user profile for userID: ${userId}`);

    const response = await fetch(`http://127.0.0.1:8079/viewprofil/${userId}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
      credentials: "include",
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Erreur lors du chargement du profil.");
    }

    return await response.json(); 
  } catch (error) {
    console.error("Error fetching user profile:", error.message);
    throw error;
  }
};


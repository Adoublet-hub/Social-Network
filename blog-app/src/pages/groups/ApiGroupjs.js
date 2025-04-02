import apiRequest  from "../profil/SearchBar"; 

const API_BASE_URL = "http://127.0.0.1:8079";


  export const handleCreateGroup = async (groupData) => {
    try {
      const endpoint = "/create_group";
      const method = "POST";
  
      const response = await apiRequest(endpoint, method, groupData);
  
      console.log("Group created successfully:", response);
      return response; 
    } catch (error) {
      console.error("Error creating group:", error);
      throw error;
    }
  };
  

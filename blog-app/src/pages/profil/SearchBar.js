import React, { useState, useEffect } from "react";
import { useRouter } from "next/router";

const API_BASE_URL = "http://127.0.0.1:8079";


export const apiRequest = async (endpoint, method = "GET", body = null) => {
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
    console.log("Response status:", response.status);

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "API request failed");
    }
    return await response.json();
  } catch (error) {
    console.error(`Error in ${endpoint}:`, error);
    if (error.name === "TypeError") {
      console.error("Network error or backend not reachable:", error);
    }
    throw error;
  }
};


export const SearchBar = ({ onSearch }) => {
  const [query, setQuery] = useState("");

  const handleInputChange = (e) => {
    const newQuery = e.target.value;
    setQuery(newQuery);
    onSearch(newQuery); 
  };

  return (
    <div className="flex items-center space-x-2 bg-gray-800 p-2 rounded-lg">
      <input
        type="text"
        value={query}
        onChange={handleInputChange}
        placeholder="Rechercher des utilisateurs..."
        className="w-full bg-gray-700 text-white px-4 py-2 rounded-lg outline-none"
      />
    </div>
  );
};

export const UserCard = ({ user, onToggleFollow }) => {
  const [isFollowing, setIsFollowing] = useState(user.isFollowing);
  const router = useRouter();

  const handleViewProfile = () => {
    router.push(`/viewprofil/${user.id}`);
  };

  const handleToggleFollow = async (event) => {
    event.stopPropagation();
    const success = await onToggleFollow(user.id);
    if (success) {
      setIsFollowing(!isFollowing);
    }
  };

  return (
    <div 
      onClick={handleViewProfile}
      className="p-4 bg-gray-700 rounded-lg shadow-md text-center cursor-pointer hover:bg-gray-600 transition"
    >
      <img
        src={user.avatar || "/avatar.png"}
        alt={user.username}
        className="w-16 h-16 rounded-full mx-auto mb-4"
      />
      <h2 className="text-lg font-semibold text-cyan-300">{user.username}</h2>
      
      <button
        onClick={handleToggleFollow}
        className={`w-full mt-4 py-2 rounded-lg shadow-md ${
          isFollowing ? "bg-red-500 hover:bg-red-400" : "bg-cyan-600 hover:bg-cyan-500"
        } text-white`}
      >
        {isFollowing ? "Se désabonner" : "Suivre"}
      </button>
    </div>
  );
};





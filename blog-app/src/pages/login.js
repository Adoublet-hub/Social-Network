import React, { useState, useEffect } from "react";
import { useRouter } from "next/router";
import AOS from "aos";
import "aos/dist/aos.css";

export default function AuthPage() {
  const [isLogin, setIsLogin] = useState(true); 
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [formData, setFormData] = useState({
    username: "",
    age: 0,
    email: "",
    password: "",
    firstName: "",
    lastName: "",
    gender: "",
    dateOfBirth: "",
    avatar: "",
    bio: "",
    phoneNumber: "",
    address: "",
    isPrivate: false,
  });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const router = useRouter();

  useEffect(() => {
    AOS.init({ duration: 1000 });
  }, []);

  const validateEmail = (email) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  const validatePassword = (password) => /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[\W_]).{8,}$/.test(password);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: name === "age" ? parseInt(value, 10) || 0 : value,
    }));
  };
  

  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (!validateEmail(email)) {
      setError("Veuillez entrer un email valide.");
      return;
    }

    try {
      const response = await fetch("http://127.0.0.1:8079/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const contentType = response.headers.get("content-type");
        if (contentType && contentType.includes("application/json")) {
          const errorResponse = await response.json();
          setError(`Erreur: ${errorResponse.error || "Échec de la connexion, vérifiez vos identifiants."}`);
        } else {
          setError("Une erreur est survenue lors de la connexion. Veuillez réessayer.");
        }
      } else {
        const data = await response.json();
        localStorage.setItem("authToken", data.token);
        setSuccess("Connexion réussie !");
        setEmail("");     
        setPassword("");   
        setTimeout(() => {
          router.push("/");
        }, 1000);
      }
    } catch (networkError) {
      setError("Erreur réseau, veuillez vérifier votre connexion et réessayer.");
    }
  };

  const handleRegisterSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");
  
    if (!validateEmail(formData.email)) {
      setError("Veuillez entrer un email valide.");
      return;
    }
  
    if (!validatePassword(formData.password)) {
      setError(
        "Le mot de passe doit comporter au moins 8 caractères, inclure des lettres majuscules et minuscules, un chiffre et un symbole spécial."
      );
      return;
    }
  
    try {
      const response = await fetch("http://127.0.0.1:8079/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: JSON.stringify(formData),
      });
  
      const contentType = response.headers.get("content-type");
      const responseBody = await response.text();
  
      if (!response.ok) {
        if (contentType && contentType.includes("application/json")) {
          const errorResponse = JSON.parse(responseBody);
          setError(`Erreur: ${errorResponse.error || "Échec de la création de l'utilisateur."}`);
        } else {
          setError("Une erreur est survenue lors de la connexion. Veuillez réessayer.");
        }
      } else {
        const data = JSON.parse(responseBody);
        console.log("Utilisateur enregistré avec succès :", data);
  
        if (data && data.user && data.user.id && data.user.id !== "00000000-0000-0000-0000-000000000000") {
          setSuccess("Inscription réussie, connectez vous!");
          setFormData({
            username: "",
            age: 0,
            email: "",
            password: "",
            firstName: "",
            lastName: "",
            gender: "",
            dateOfBirth: "",
            avatar: "",
            bio: "",
            phoneNumber: "",
            address: "",
            isPrivate: false,
          });
          
          setTimeout(() => {
            setSuccess(""); 
            setIsLogin(true);
            router.push("/login");
          }, 2000);
        } else {
          setError("Erreur lors de la création de l'utilisateur. Données invalides reçues.");
        }
      }
    } catch (error) {
      console.error("Erreur lors de la requête de connexion:", error);
      setError("Erreur réseau, veuillez vérifier votre connexion et réessayer.");
    }
  };
  

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 text-gray-100 p-4 sm:p-10">
      <div data-aos="fade-up" className="w-full max-w-lg bg-gray-800 rounded-xl shadow-lg p-8 sm:p-12 transition-all transform hover:scale-105">
        <div className="text-center mb-8">
          <button
            onClick={() => setIsLogin(!isLogin)}
            className="text-indigo-400 hover:underline focus:outline-none"
          >
            {isLogin ? "Vous n'avez pas de compte ? Créez un compte" : "Déjà un compte ? Connectez-vous"}
          </button>
        </div>

        <h2 className="text-3xl sm:text-4xl font-extrabold text-center mb-6 animate-pulse">
          {isLogin ? "Connexion" : "Inscription"}
        </h2>
        
        {error && <p className="text-red-500 bg-red-100 p-3 rounded-lg text-center mb-4">{error}</p>}
        {success && <p className="text-green-500 bg-green-100 p-3 rounded-lg text-center mb-4">{success}</p>}

        <form className="space-y-6" onSubmit={isLogin ? handleLoginSubmit : handleRegisterSubmit}>
          {!isLogin && (
            <>
              <div>
                <label htmlFor="username" className="block text-sm font-medium text-gray-400">Nom d'utilisateur</label>
                <input
                  id="username"
                  name="username"
                  type="text"
                  value={formData.username}
                  onChange={handleChange}
                  className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
                  required
                />
              </div>
              <div>
                <label htmlFor="gender" className="block text-sm font-medium text-gray-400">Genre</label>
                <select
                  id="gender"
                  name="gender"
                  value={formData.gender}
                  onChange={handleChange}
                  className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
                  required
                >
                  <option value="">Sélectionnez</option>
                  <option value="Homme">Homme</option>
                  <option value="Femme">Femme</option>
                  <option value="Autre">Autre</option>
                </select>
              </div>
            </>
          )}
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-400">Adresse email</label>
            <input
              id="email"
              name="email"
              type="email"
              value={isLogin ? email : formData.email}
              onChange={(e) => (isLogin ? setEmail(e.target.value) : handleChange(e))}
              className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
              required
            />
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-400">Mot de passe</label>
            <input
              id="password"
              name="password"
              type="password"
              value={isLogin ? password : formData.password}
              onChange={(e) => (isLogin ? setPassword(e.target.value) : handleChange(e))}
              className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
              required
            />
            {!isLogin && (
              <p className="text-xs text-gray-400 mt-1">
                Doit contenir au moins 8 caractères, une majuscule, un chiffre et un symbole.
              </p>
            )}
          </div>
          <button
            type="submit"
            className="w-full py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg shadow-lg transform transition-transform duration-300 hover:scale-105"
          >
            {isLogin ? "Connexion" : "S'inscrire"}
          </button>
        </form>
      </div>
    </div>
  );
}

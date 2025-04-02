import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import HomePage from '../pages/home';

export default function Index() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [serverError, setServerError] = useState(false);

  useEffect(() => {
    const checkAuthAndServer = async () => {
      const token = localStorage.getItem("authToken");
      if (!token) {
        router.replace("/login");
        return;
      }
  
      const controller = new AbortController();
      const signal = controller.signal;
  
      try {
        const response = await fetch("http://127.0.0.1:8079/verify_token", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          signal,
        });
  
        if (response.ok) {
          setIsAuthenticated(true);
        } else if (response.status === 401) {
          localStorage.removeItem("authToken"); 
          router.replace("/login"); 
        } else {
          router.replace("/login");
        }
      } catch (error) {
        if (Error.name !== 'AbortError') {
          console.error("Server is down or unreachable", error);
          setServerError(true);
        }
      } finally {
        setLoading(false);
      }
  
      return () => controller.abort();
    };
  
    checkAuthAndServer();
  }, [router]);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <div className="flex flex-col items-center">
          <div className="animate-spin rounded-full h-16 w-16 border-t-4 border-blue-500"></div>
          <p className="mt-4 text-blue-600">Chargement...</p>
        </div>
      </div>
    );
  }

  if (serverError) {
    return (
      <div className="flex justify-center items-center h-screen">
        <p className="text-center text-red-600 animate-pulse">
          Le serveur est actuellement indisponible. Veuillez réessayer plus tard.
        </p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null; 
  }

  return <HomePage />;
}
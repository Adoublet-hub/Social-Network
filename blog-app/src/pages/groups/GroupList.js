import { useState, useEffect } from "react";
import { useRouter } from "next/router";
import { motion } from "framer-motion";
import { IconArrowLeft } from "@tabler/icons-react";


export default function GroupListPage() {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isMember, setIsMember] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const fetchGroups = async () => {
      try {
        setLoading(true);
        const response = await fetch(
          `http://127.0.0.1:8079/list_group?page=${page}&limit=10`,
          {
            headers: {
              Authorization: `Bearer ${localStorage.getItem("authToken")}`,
            },
          }
        );

        if (response.status === 403) {
          setIsMember(false); 
          return;
        }

        if (!response.ok) throw new Error("Erreur lors du chargement des groupes");

        const data = await response.json();

        if (!Array.isArray(data)) {
          throw new Error("Réponse inattendue du serveur");
        }

        setGroups((prev) => {
          const combined = [...prev, ...data];
          return combined.filter(
            (group, index, self) => self.findIndex((g) => g.id === group.id) === index
          );
        });

        setHasMore(data.length > 0);
      } catch (err) {
        setError("Impossible de récupérer les groupes.");
      } finally {
        setLoading(false);
      }
    };

    fetchGroups();
  }, [page]);

  const navigateToGroup = (id) => {
    router.push(`/groups/${id}`);
  };

  // Load the next page
  const handleNextPage = () => {
    if (hasMore) setPage((prev) => prev + 1);
  };

  if (!isMember) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900 text-gray-200">
        <p className="text-center text-red-500">Vous devez être membre d'un groupe pour voir cette page.</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-gray-200">
      <header className="w-full px-4 py-2 bg-gray-800 shadow-md flex items-center">
        <motion.button
          onClick={() => router.push("/")}
          whileHover={{ scale: 1.05 }}
          className="flex items-center text-cyan-400 hover:text-cyan-300 transition duration-200"
        >
          <IconArrowLeft className="h-5 w-5 mr-2" />
          Retour à l'accueil
        </motion.button>
      </header>

      <div className="max-w-4xl mx-auto my-8 px-6 py-8 bg-gray-900 rounded-xl shadow-lg">
        <motion.h2
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-3xl font-semibold text-center text-cyan-400"
        >
          Groupes
        </motion.h2>

        {error && <p className="text-red-500 text-center">{error}</p>}

        <div className="space-y-6">
          {groups.map((group) => (
            <div
              key={group.id}
              className="bg-gray-800 p-6 rounded-lg shadow-md hover:bg-gray-700 cursor-pointer"
              onClick={() => navigateToGroup(group.id)}
            >
              <h3 className="text-2xl font-bold text-cyan-300">{group.name}</h3>
              <p className="text-gray-400">{group.description}</p>
            </div>
          ))}
        </div>

        {hasMore && !loading && (
          <button
            onClick={handleNextPage}
            className="w-full mt-6 bg-cyan-600 text-white py-3 rounded-lg hover:bg-cyan-500 transition"
          >
            Charger plus
          </button>
        )}

        {loading && <p className="text-center mt-4">Chargement...</p>}
      </div>
    </div>
  );
}

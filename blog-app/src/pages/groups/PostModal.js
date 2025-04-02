import GroupPosts from "./GroupPosts";

export default function PostModal({ groupId, onClose }) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-gray-800 p-6 rounded-lg shadow-lg w-96 max-h-[80vh] overflow-y-auto">
        <button
          onClick={onClose}
          className="text-red-400 hover:text-red-300 float-right"
        >
          ✖
        </button>
        <h2 className="text-2xl text-white mb-4">Posts du Groupe</h2>
        <GroupPosts groupId={groupId} />
      </div>
    </div>
  );
}

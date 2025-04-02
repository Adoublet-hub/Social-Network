import { useState } from "react";
import EmojiPicker from "./EmojiPicker";
import { IconSend, IconMoodSmile, IconPhoto, IconX } from "@tabler/icons-react";

export default function MessageInputBar({ onSendMessage, onSendImage }) {
  const [message, setMessage] = useState("");
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [selectedImage, setSelectedImage] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);

  const handleSendMessage = () => {
    if (message.trim()) {
      onSendMessage(message);
      setMessage("");
    }
  };

  const handleEmojiSelect = (emoji) => {
    if (emoji && (emoji.native || emoji.emoji)) {
      const emojiChar = emoji.native || emoji.emoji;
      setMessage((prev) => prev + emojiChar);
    } else {
      console.error("Emoji sélectionné invalide :", emoji);
    }
    setShowEmojiPicker(false);
  };
  
  

  const handleImageChange = (event) => {
    const file = event.target.files[0];
    if (file) {
      setSelectedImage(file);
      setPreviewUrl(URL.createObjectURL(file));
    }
  };

  const handleCancelImage = () => {
    setSelectedImage(null);
    setPreviewUrl(null);
  };

  const handleSendImage = () => {
    if (selectedImage) {
      if (typeof onSendImage === "function") {
        onSendImage(selectedImage);
      } else {
        console.warn("onSendImage n'est pas défini !");
      }
      setSelectedImage(null);
      setPreviewUrl(null);
    }
  };
  

  return (
    <div className="p-4 bg-gray-800 border-t border-gray-700">
      {previewUrl && (
        <div className="flex items-center mb-4 space-x-4">
          <img
            src={previewUrl}
            alt="Prévisualisation"
            className="w-20 h-20 object-cover rounded-lg"
          />
          <button
            onClick={handleCancelImage}
            className="p-2 bg-red-500 text-white rounded-lg hover:bg-red-400"
          >
            <IconX />
          </button>
        </div>
      )}

      <div className="flex items-center space-x-2">
        <div className="relative">
          <button
            onClick={() => setShowEmojiPicker(!showEmojiPicker)}
            className="p-2 bg-gray-700 rounded-lg text-gray-300"
          >
            <IconMoodSmile />
          </button>
          {showEmojiPicker && (
            <div className="absolute bottom-12 left-0 z-10">
              <EmojiPicker onEmojiSelect={handleEmojiSelect} />
            </div>
          )}
        </div>

        <label className="p-2 bg-gray-700 rounded-lg text-gray-300 cursor-pointer">
          <IconPhoto />
          <input
            type="file"
            accept="image/*"
            className="hidden"
            onChange={handleImageChange}
          />
        </label>

        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Écrivez un message..."
          className="flex-1 p-2 bg-gray-700 rounded-lg text-gray-300"
        />

        <button
          onClick={selectedImage ? handleSendImage : handleSendMessage}
          className="p-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-500"
        >
          <IconSend />
        </button>
      </div>
    </div>
  );
}

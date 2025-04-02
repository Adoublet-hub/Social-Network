import Picker from "emoji-picker-react";

export default function EmojiPicker({ onEmojiSelect }) {
  const handleEmojiClick = (emojiData, event) => {
    console.log("Emoji sélectionné :", emojiData);  
    onEmojiSelect(emojiData);
  };

  return <Picker onEmojiClick={handleEmojiClick} />;
}

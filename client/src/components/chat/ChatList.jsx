import React from 'react';

// Placeholder for ChatList until chatStore is fully implemented
export default function ChatList() {
  const dummyChats = [
    { id: 1, name: "Samet", discriminator: "0432", lastMessage: "Naber?", unread: 2, status: "online" },
    { id: 2, name: "Ahmet", discriminator: "1234", lastMessage: "Tamamdır, görüşürüz.", unread: 0, status: "offline" }
  ];

  return (
    <div className="flex flex-col p-2 space-y-1">
      {dummyChats.map(chat => (
        <button key={chat.id} className="flex items-center gap-3 p-3 rounded-xl hover:bg-white/5 transition-colors text-left relative group">
          <div className="relative">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary/40 to-accent/40 flex items-center justify-center font-bold text-white">
              {chat.name.charAt(0)}
            </div>
            <span className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-[#1E1E24] ${chat.status === 'online' ? 'bg-green-500' : 'bg-gray-500'}`}></span>
          </div>
          
          <div className="flex-1 overflow-hidden">
            <div className="flex justify-between items-baseline">
              <span className="font-medium text-sm truncate">{chat.name}</span>
              <span className="text-[10px] text-white/30">12:34</span>
            </div>
            <p className="text-xs text-white/50 truncate group-hover:text-white/70 transition-colors">
              {chat.lastMessage}
            </p>
          </div>

          {chat.unread > 0 && (
            <span className="bg-primary text-white text-[10px] font-bold px-2 py-0.5 rounded-full">
              {chat.unread}
            </span>
          )}
        </button>
      ))}
    </div>
  );
}

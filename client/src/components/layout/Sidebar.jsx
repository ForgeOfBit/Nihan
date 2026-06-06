import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import useAuthStore from '../../stores/authStore';
import ChatList from '../chat/ChatList';
import Avatar from '../common/Avatar';
import UserTag from '../common/UserTag';

export default function Sidebar() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('chats'); // 'chats' or 'friends'

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="w-80 h-full flex flex-col bg-surface/50 backdrop-blur-xl border-r border-white/5">
      {/* Header */}
      <div className="p-4 border-b border-white/5">
        <h2 className="text-xl font-bold bg-gradient-to-r from-primary to-accent bg-clip-text text-transparent">Nihan</h2>
      </div>

      {/* Tabs */}
      <div className="flex p-2 gap-2 border-b border-white/5">
        <button 
          onClick={() => setActiveTab('chats')}
          className={`flex-1 py-1.5 px-3 rounded-lg text-sm font-medium transition-all ${activeTab === 'chats' ? 'bg-white/10 text-white' : 'text-white/50 hover:text-white/80 hover:bg-white/5'}`}
        >
          Sohbetler
        </button>
        <button 
          onClick={() => setActiveTab('friends')}
          className={`flex-1 py-1.5 px-3 rounded-lg text-sm font-medium transition-all ${activeTab === 'friends' ? 'bg-white/10 text-white' : 'text-white/50 hover:text-white/80 hover:bg-white/5'}`}
        >
          Arkadaşlar
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === 'chats' ? (
          <ChatList />
        ) : (
          <div className="p-4 text-center text-sm text-white/40">Arkadaş listesi yakında...</div>
        )}
      </div>

      {/* User Profile Bar */}
      <div className="p-3 bg-black/20 flex items-center justify-between border-t border-white/5">
        <Link to="/settings" className="flex items-center gap-3 hover:bg-white/5 p-2 rounded-xl transition-colors flex-1 overflow-hidden">
          <Avatar src={user?.avatarUrl} alt={user?.username} status={user?.status} size="sm" />
          <div className="flex flex-col truncate">
            <span className="text-sm font-semibold text-white truncate">{user?.displayName || user?.username}</span>
            <UserTag username={user?.username} discriminator={user?.discriminator} className="text-xs text-white/50" />
          </div>
        </Link>
        <button 
          onClick={handleLogout}
          className="p-2 text-white/40 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition-all ml-2"
          title="Çıkış Yap"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 9.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 9H7a1 1 0 100 2h7.586l-1.293 1.293z" clipRule="evenodd" />
          </svg>
        </button>
      </div>
    </div>
  );
}

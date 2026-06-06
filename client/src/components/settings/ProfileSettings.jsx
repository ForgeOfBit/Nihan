import React, { useState } from 'react';
import useAuthStore from '../../stores/authStore';

export default function ProfileSettings() {
  const { user } = useAuthStore();
  const [displayName, setDisplayName] = useState(user?.displayName || '');
  const [bio, setBio] = useState(user?.bio || '');
  const [loading, setLoading] = useState(false);

  const handleSave = async (e) => {
    e.preventDefault();
    setLoading(true);
    // API call to update profile will go here
    setTimeout(() => setLoading(false), 1000);
  };

  return (
    <div className="bg-surface/50 backdrop-blur-xl border border-white/5 rounded-2xl p-6">
      <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-primary" viewBox="0 0 20 20" fill="currentColor">
          <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
        </svg>
        Profil Bilgileri
      </h2>
      
      <form onSubmit={handleSave} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-white/70 mb-1">Görünen Ad</label>
          <input 
            type="text" 
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className="w-full px-4 py-2.5 rounded-xl bg-black/20 border border-white/10 text-white focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/20"
            placeholder={user?.username}
          />
        </div>
        
        <div>
          <label className="block text-sm font-medium text-white/70 mb-1">Hakkımda</label>
          <textarea 
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            className="w-full px-4 py-2.5 rounded-xl bg-black/20 border border-white/10 text-white focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/20 resize-none h-24"
            placeholder="Kendinizden bahsedin..."
          />
        </div>

        <div className="flex justify-end pt-2">
          <button 
            type="submit" 
            disabled={loading}
            className="px-6 py-2.5 rounded-xl bg-white/10 hover:bg-white/20 text-white font-medium transition-all"
          >
            {loading ? 'Kaydediliyor...' : 'Değişiklikleri Kaydet'}
          </button>
        </div>
      </form>
    </div>
  );
}

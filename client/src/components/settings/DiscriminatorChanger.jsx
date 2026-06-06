import React, { useState } from 'react';
import useAuthStore from '../../stores/authStore';

export default function DiscriminatorChanger() {
  const { user } = useAuthStore();
  const [newDisc, setNewDisc] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');

  const handleUpdate = async (e) => {
    e.preventDefault();
    setLoading(true);
    setMessage('');
    
    // Simulate API Call
    setTimeout(() => {
      setMessage('Discriminator başarıyla güncellendi!');
      setLoading(false);
    }, 1500);
  };

  return (
    <div className="bg-gradient-to-br from-purple-900/40 to-blue-900/40 backdrop-blur-xl border border-primary/20 rounded-2xl p-6 relative overflow-hidden">
      {/* Premium badge */}
      <div className="absolute top-0 right-0 bg-gradient-to-r from-yellow-400 to-yellow-600 text-black text-xs font-bold px-3 py-1 rounded-bl-lg shadow-lg">
        PREMIUM
      </div>

      <h2 className="text-xl font-bold text-white mb-2 flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
        </svg>
        Özel Etiket (Discriminator)
      </h2>
      
      <p className="text-white/60 text-sm mb-6 max-w-2xl">
        Premium üye olarak kullanıcı adınızın sonundaki 4 haneli sayıyı istediğiniz gibi değiştirebilirsiniz. 
        Kullanıcı adı ve etiket kombinasyonunun benzersiz olması gerektiğini unutmayın.
      </p>
      
      <div className="flex items-center gap-4 mb-6">
        <div className="bg-black/30 px-4 py-3 rounded-xl border border-white/10 flex items-center">
          <span className="font-semibold text-white/90 text-lg">{user?.username}</span>
          <span className="text-primary font-bold text-lg ml-0.5">#{user?.discriminator || '0000'}</span>
        </div>
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-white/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
        </svg>
      </div>

      <form onSubmit={handleUpdate} className="flex gap-3">
        <div className="relative flex-1 max-w-xs">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40 font-bold text-lg">#</span>
          <input 
            type="text" 
            maxLength={4}
            pattern="\d{4}"
            value={newDisc}
            onChange={(e) => setNewDisc(e.target.value.replace(/\D/g, ''))}
            className="w-full pl-8 pr-4 py-3 rounded-xl bg-black/20 border border-primary/30 text-white font-bold tracking-widest focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/20"
            placeholder="0001"
            disabled={!user?.isPremium}
          />
        </div>
        
        <button 
          type="submit" 
          disabled={!user?.isPremium || newDisc.length !== 4 || loading}
          className="px-6 py-3 rounded-xl bg-gradient-to-r from-primary to-accent text-white font-medium shadow-lg shadow-primary/20 hover:shadow-primary/40 disabled:opacity-50 transition-all"
        >
          {loading ? 'Güncelleniyor...' : 'Güncelle'}
        </button>
      </form>

      {message && (
        <div className="mt-4 text-green-400 text-sm font-medium">
          {message}
        </div>
      )}

      {!user?.isPremium && (
        <div className="mt-6 pt-4 border-t border-white/10">
          <button className="text-sm font-medium text-yellow-400 hover:text-yellow-300 flex items-center gap-1 transition-colors">
            Premium'a Yükselt ve Özelleştirmeye Başla
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M12.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-2.293-2.293a1 1 0 010-1.414z" clipRule="evenodd" /></svg>
          </button>
        </div>
      )}
    </div>
  );
}

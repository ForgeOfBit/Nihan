import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import useAuthStore from '../stores/authStore';

export default function RegisterPage() {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { register, loading } = useAuthStore();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    
    // Basic validation
    if (username.length < 3 || username.length > 32) {
      setError('Kullanıcı adı 3-32 karakter arasında olmalıdır.');
      return;
    }

    try {
      await register(username, email, password);
      navigate('/');
    } catch (err) {
      setError(err.message || 'Kayıt başarısız oldu.');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background relative overflow-hidden">
      {/* Background Animated Gradients */}
      <div className="absolute top-[-20%] right-[-10%] w-[50%] h-[50%] rounded-full bg-accent/20 blur-[120px] animate-pulse-slow"></div>
      <div className="absolute bottom-[-20%] left-[-10%] w-[50%] h-[50%] rounded-full bg-primary/20 blur-[120px] animate-pulse-slow" style={{ animationDelay: '2s' }}></div>
      
      <div className="relative z-10 w-full max-w-md p-8 rounded-2xl bg-surface/50 backdrop-blur-xl border border-white/10 shadow-2xl">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">Aramıza Katıl</h1>
          <p className="text-white/60">Güvenli ve gizli mesajlaşmaya başlayın.</p>
        </div>

        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm font-medium text-white/80 mb-2">Kullanıcı Adı</label>
            <input 
              type="text" 
              required
              value={username}
              onChange={(e) => setUsername(e.target.value.replace(/[^a-zA-Z0-9_]/g, ''))}
              className="w-full px-4 py-3 rounded-xl bg-black/20 border border-white/10 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-white/30"
              placeholder="Sadece harf, rakam ve _"
            />
            {username && (
              <p className="mt-2 text-xs text-white/50 flex items-center">
                Görünümünüz: <span className="ml-1 font-medium text-primary">{username}<span className="text-white/30">#XXXX</span></span>
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-white/80 mb-2">E-posta</label>
            <input 
              type="email" 
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-black/20 border border-white/10 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-white/30"
              placeholder="ornek@email.com"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-white/80 mb-2">Şifre</label>
            <input 
              type="password" 
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-black/20 border border-white/10 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-white/30"
              placeholder="En az 8 karakter"
              minLength={8}
            />
          </div>

          <button 
            type="submit" 
            disabled={loading}
            className="w-full py-3 px-4 mt-2 rounded-xl bg-gradient-to-r from-primary to-accent text-white font-semibold shadow-lg shadow-primary/25 hover:shadow-primary/40 transform hover:-translate-y-0.5 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
          >
            {loading ? 'Kayıt Yapılıyor...' : 'Hesap Oluştur'}
          </button>
        </form>

        <p className="mt-8 text-center text-sm text-white/50">
          Zaten hesabınız var mı? <Link to="/login" className="text-primary hover:text-accent transition-colors font-medium">Giriş Yap</Link>
        </p>
      </div>
    </div>
  );
}

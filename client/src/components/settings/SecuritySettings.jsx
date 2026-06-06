import React from 'react';

export default function SecuritySettings() {
  return (
    <div className="bg-surface/50 backdrop-blur-xl border border-white/5 rounded-2xl p-6">
      <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
          <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
        </svg>
        Güvenlik ve Şifreleme
      </h2>
      
      <div className="space-y-6">
        <div className="flex items-start gap-4 p-4 bg-green-500/10 border border-green-500/20 rounded-xl">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-green-400 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <div>
            <h3 className="font-semibold text-green-400 mb-1">Uçtan Uca Şifreleme (E2EE) Aktif</h3>
            <p className="text-sm text-green-400/80">
              Mesajlarınız cihazınızda şifrelenir ve sadece alıcı tarafından çözülebilir. 
              Sunucularımız dahil hiç kimse mesajlarınızı okuyamaz.
            </p>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="p-4 bg-black/20 rounded-xl border border-white/5">
            <h4 className="text-xs font-semibold text-white/50 uppercase tracking-wider mb-2">Kimlik Anahtarı</h4>
            <div className="font-mono text-sm text-white/90 break-all bg-black/40 p-2 rounded truncate" title="X25519 ve Ed25519 anahtarları tarayıcınızın güvenli deposunda (IndexedDB) tutulmaktadır.">
              X25519 / Ed25519 (Cihazda Saklanıyor)
            </div>
            <p className="text-xs text-white/40 mt-2">Bu cihaza özel şifreleme anahtarınız.</p>
          </div>
          
          <div className="p-4 bg-black/20 rounded-xl border border-white/5">
            <h4 className="text-xs font-semibold text-white/50 uppercase tracking-wider mb-2">Tek Kullanımlık Anahtarlar</h4>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold text-white">48</span>
              <span className="text-sm text-white/50">/ 100</span>
            </div>
            <p className="text-xs text-white/40 mt-1">Sunucuda bekleyen pre-key sayısı.</p>
            <button className="mt-2 text-xs font-medium text-primary hover:text-accent transition-colors">
              Anahtarları Yenile
            </button>
          </div>
        </div>
        
        <div className="pt-4 border-t border-white/10">
          <button className="px-4 py-2 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20 text-sm font-medium transition-all">
            Şifreleme Anahtarlarını Sıfırla
          </button>
          <p className="text-xs text-white/40 mt-2">
            Dikkat: Bu işlem mevcut cihazınızdaki eski mesajların okunamamasına neden olabilir.
          </p>
        </div>
      </div>
    </div>
  );
}

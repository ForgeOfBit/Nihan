import React from 'react';

export default function Avatar({ src, alt, status = 'offline', size = 'md' }) {
  const sizeClasses = {
    sm: 'w-8 h-8',
    md: 'w-10 h-10',
    lg: 'w-12 h-12',
    xl: 'w-24 h-24'
  };

  const statusColors = {
    online: 'bg-green-500',
    idle: 'bg-yellow-500',
    dnd: 'bg-red-500',
    offline: 'bg-gray-500'
  };

  return (
    <div className={`relative ${sizeClasses[size]}`}>
      {src ? (
        <img 
          src={src} 
          alt={alt} 
          className="w-full h-full rounded-full object-cover bg-black/20"
        />
      ) : (
        <div className="w-full h-full rounded-full bg-gradient-to-br from-primary/50 to-accent/50 flex items-center justify-center text-white font-bold text-lg uppercase">
          {alt?.charAt(0) || '?'}
        </div>
      )}
      
      {/* Status Indicator */}
      <span 
        className={`absolute bottom-0 right-0 block rounded-full ring-2 ring-background ${statusColors[status]}`}
        style={{ width: '28%', height: '28%' }}
      />
    </div>
  );
}

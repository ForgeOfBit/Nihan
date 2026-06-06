import React from 'react';

export default function UserTag({ username, discriminator, className = "" }) {
  if (!username) return null;
  
  return (
    <span className={`inline-flex items-center ${className}`}>
      <span className="font-medium">{username}</span>
      <span className="opacity-50 text-[0.9em]">#{discriminator || '0000'}</span>
    </span>
  );
}

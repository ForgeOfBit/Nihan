import React from 'react';
import ChatLayout from '../components/layout/ChatLayout';
import ProfileSettings from '../components/settings/ProfileSettings';
import DiscriminatorChanger from '../components/settings/DiscriminatorChanger';
import SecuritySettings from '../components/settings/SecuritySettings';

export default function SettingsPage() {
  return (
    <ChatLayout>
      <div className="flex-1 overflow-y-auto p-8">
        <div className="max-w-3xl mx-auto space-y-8">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">Ayarlar</h1>
            <p className="text-white/50">Profilinizi, güvenliğinizi ve premium özelliklerinizi yönetin.</p>
          </div>
          
          <ProfileSettings />
          <DiscriminatorChanger />
          <SecuritySettings />
        </div>
      </div>
    </ChatLayout>
  );
}

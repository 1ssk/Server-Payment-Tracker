import { useState, useEffect } from 'react';
import { SMTPSettings } from '../types/server';
import { Mail, Save, X, CheckCircle, AlertCircle } from 'lucide-react';
import { apiGetSMTPSettings, apiUpdateSMTPSettings, apiSendTestEmail } from '../utils/api';

interface SettingsPanelProps {
  onClose: () => void;
}

export function SettingsPanel({ onClose }: SettingsPanelProps) {
  const [settings, setSettings] = useState<SMTPSettings>({
    host: '',
    port: 587,
    username: '',
    password: '',
    from: '',
    to: '',
    enabled: false,
    reminderDaysBefore: 10
  });
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await apiGetSMTPSettings();
        setSettings(data);
      } catch (error) {
        console.error('Error loading SMTP settings:', error);
      }
    };
    load();
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const saved = await apiUpdateSMTPSettings(settings);
      setSettings(saved);
      setSaveMessage({ type: 'success', text: 'Настройки SMTP успешно сохранены' });
      
      setTimeout(() => {
        setSaveMessage(null);
      }, 3000);
    } catch (error) {
      console.error('Error saving SMTP settings:', error);
      setSaveMessage({ type: 'error', text: 'Ошибка при сохранении настроек' });
    }
  };

  const handleTestEmail = async () => {
    if (!settings.enabled) {
      setSaveMessage({ type: 'error', text: 'Включите уведомления для отправки тестового письма' });
      return;
    }
    try {
      await apiSendTestEmail();
      setSaveMessage({ type: 'success', text: 'Тестовое письмо отправлено на указанный адрес.' });
    } catch (e) {
      setSaveMessage({ type: 'error', text: 'Не удалось отправить письмо. Проверьте настройки SMTP.' });
    }
    setTimeout(() => setSaveMessage(null), 5000);
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-indigo-50 text-indigo-600 rounded-lg">
              <Mail className="w-5 h-5" />
            </div>
            <h2 className="text-xl font-semibold text-gray-900">Настройки Email-уведомлений</h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSave} className="p-6 space-y-5">
          {saveMessage && (
            <div className={`p-4 rounded-lg flex items-center gap-3 ${
              saveMessage.type === 'success' 
                ? 'bg-green-50 border border-green-200 text-green-700'
                : 'bg-red-50 border border-red-200 text-red-700'
            }`}>
              {saveMessage.type === 'success' ? (
                <CheckCircle className="w-5 h-5 flex-shrink-0" />
              ) : (
                <AlertCircle className="w-5 h-5 flex-shrink-0" />
              )}
              <p className="text-sm">{saveMessage.text}</p>
            </div>
          )}

          <div className="flex items-center justify-between p-4 bg-indigo-50 rounded-lg border border-indigo-200">
            <div>
              <h3 className="font-medium text-gray-900">Включить email-уведомления</h3>
              <p className="text-sm text-gray-600 mt-1">
                Получать уведомления о предстоящих платежах на почту
              </p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.enabled}
                onChange={(e) => setSettings({ ...settings, enabled: e.target.checked })}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600"></div>
            </label>
          </div>

          <div className="space-y-4">
            <h3 className="font-medium text-gray-900">Параметры SMTP-сервера</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  SMTP хост *
                </label>
                <input
                  type="text"
                  required={settings.enabled}
                  value={settings.host}
                  onChange={(e) => setSettings({ ...settings, host: e.target.value })}
                  className="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                  placeholder="smtp.gmail.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Порт *
                </label>
                <input
                  type="number"
                  required={settings.enabled}
                  value={settings.port}
                  onChange={(e) => setSettings({ ...settings, port: parseInt(e.target.value) || 587 })}
                  className="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                  placeholder="587"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Имя пользователя *
                </label>
                <input
                  type="text"
                  required={settings.enabled}
                  value={settings.username}
                  onChange={(e) => setSettings({ ...settings, username: e.target.value })}
                  className="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                  placeholder="user@example.com"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Пароль *
                </label>
                <input
                  type="password"
                  required={settings.enabled}
                  value={settings.password}
                  onChange={(e) => setSettings({ ...settings, password: e.target.value })}
                  className="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                  placeholder="••••••••••••"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  От кого (From) *
                </label>
                <input
                  type="email"
                  required={settings.enabled}
                  value={settings.from}
                  onChange={(e) => setSettings({ ...settings, from: e.target.value })}
                  className="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                  placeholder="noreply@example.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Кому (To) *
                </label>
                <input
                  type="email"
                  required={settings.enabled}
                  value={settings.to}
                  onChange={(e) => setSettings({ ...settings, to: e.target.value })}
                  className="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                  placeholder="email@example.com"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Напоминание за N дней до срока оплаты
                </label>
                <input
                  type="number"
                  min={1}
                  max={60}
                  value={settings.reminderDaysBefore || 10}
                  onChange={(e) => setSettings({ ...settings, reminderDaysBefore: parseInt(e.target.value, 10) || 10 })}
                  className="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                />
                <p className="text-xs text-gray-500 mt-1">По умолчанию письмо приходит за 10 дней до даты платежа.</p>
              </div>
            </div>
          </div>

          <div className="flex gap-3 pt-4 border-t border-gray-200">
            <button
              type="button"
              onClick={handleTestEmail}
              className="px-4 py-2.5 text-indigo-700 bg-indigo-50 border border-indigo-200 rounded-lg hover:bg-indigo-100 transition-colors"
            >
              Отправить тестовое письмо
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2.5 text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors flex items-center justify-center gap-2"
            >
              <Save className="w-5 h-5" />
              Сохранить настройки
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

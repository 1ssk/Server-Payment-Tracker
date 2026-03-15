import { useState } from 'react';
import { Server } from '../types/server';
import { Server as ServerIcon, MapPin, Calendar, DollarSign, Edit, Trash2, ExternalLink, AlertCircle, CheckCircle } from 'lucide-react';
import { calculateYearlyCost, getDaysUntilPayment } from '../utils/calculations';

interface ServerCardProps {
  server: Server;
  onEdit: (server: Server) => void;
  onDelete: (id: string) => void;
  onConfirmPayment: (serverId: string, paidAt: string, amount?: number) => Promise<void>;
}

export function ServerCard({ server, onEdit, onDelete, onConfirmPayment }: ServerCardProps) {
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [confirmDate, setConfirmDate] = useState(new Date().toISOString().slice(0, 10));
  const [confirmAmount, setConfirmAmount] = useState(server.monthlyCost.toString());
  const [saving, setSaving] = useState(false);
  const yearlyCost = calculateYearlyCost(server);
  const daysUntil = getDaysUntilPayment(server.nextPaymentDate);
  const isPaymentSoon = daysUntil >= 0 && daysUntil <= 7;
  const isOverdue = daysUntil < 0;

  const statusColors = {
    active: 'bg-green-100 text-green-700 border-green-200',
    inactive: 'bg-gray-100 text-gray-700 border-gray-200',
    pending: 'bg-yellow-100 text-yellow-700 border-yellow-200'
  };

  const statusLabels = {
    active: 'Активный',
    inactive: 'Неактивный',
    pending: 'Ожидание'
  };

  const billingCycleLabels = {
    monthly: 'Ежемесячно',
    quarterly: 'Ежеквартально',
    yearly: 'Ежегодно'
  };

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 hover:shadow-md transition-all">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-start gap-3 flex-1">
          <div className="p-2.5 bg-indigo-50 text-indigo-600 rounded-lg">
            <ServerIcon className="w-5 h-5" />
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-gray-900 mb-1">{server.name}</h3>
            <div className="flex items-center gap-2 flex-wrap">
              <span className={`text-xs px-2.5 py-1 rounded-full border ${statusColors[server.status]}`}>
                {statusLabels[server.status]}
              </span>
              {(isPaymentSoon || isOverdue) && server.status === 'active' && (
                <span className={`text-xs px-2.5 py-1 rounded-full flex items-center gap-1 ${
                  isOverdue 
                    ? 'bg-red-100 text-red-700 border border-red-200' 
                    : 'bg-orange-100 text-orange-700 border border-orange-200'
                }`}>
                  <AlertCircle className="w-3 h-3" />
                  {isOverdue ? 'Просрочено' : 'Скоро оплата'}
                </span>
              )}
            </div>
          </div>
        </div>
        <div className="flex gap-2">
          {server.status === 'active' && (
            <button
              onClick={() => setShowConfirmModal(true)}
              className="p-2 text-gray-500 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors"
              title="Подтвердить оплату"
            >
              <CheckCircle className="w-4 h-4" />
            </button>
          )}
          <button
            onClick={() => onEdit(server)}
            className="p-2 text-gray-500 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
            title="Редактировать"
          >
            <Edit className="w-4 h-4" />
          </button>
          <button
            onClick={() => onDelete(server.id)}
            className="p-2 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
            title="Удалить"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="space-y-2.5 mb-4">
        <div className="flex items-center gap-2 text-sm">
          <ServerIcon className="w-4 h-4 text-gray-400" />
          <span className="text-gray-600">IP:</span>
          <span className="text-gray-900 font-mono">{server.ipAddress}</span>
        </div>
        
        <div className="flex items-center gap-2 text-sm">
          <MapPin className="w-4 h-4 text-gray-400" />
          <span className="text-gray-600">Локация:</span>
          <span className="text-gray-900">{server.location}</span>
        </div>

        <div className="flex items-center gap-2 text-sm">
          <ExternalLink className="w-4 h-4 text-gray-400" />
          <span className="text-gray-600">Провайдер:</span>
          {server.providerUrl ? (
            <a 
              href={server.providerUrl} 
              target="_blank" 
              rel="noopener noreferrer"
              className="text-indigo-600 hover:text-indigo-700 hover:underline flex items-center gap-1"
            >
              {server.provider}
              <ExternalLink className="w-3 h-3" />
            </a>
          ) : (
            <span className="text-gray-900">{server.provider}</span>
          )}
        </div>

        <div className="flex items-center gap-2 text-sm">
          <Calendar className="w-4 h-4 text-gray-400" />
          <span className="text-gray-600">Следующий платёж:</span>
          <span className={`${isOverdue ? 'text-red-600 font-medium' : isPaymentSoon ? 'text-orange-600 font-medium' : 'text-gray-900'}`}>
            {new Date(server.nextPaymentDate).toLocaleDateString('ru-RU')}
            {daysUntil >= 0 && daysUntil <= 30 && ` (${daysUntil} ${daysUntil === 1 ? 'день' : daysUntil < 5 ? 'дня' : 'дней'})`}
          </span>
        </div>
      </div>

      {server.notes && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <p className="text-sm text-gray-600">{server.notes}</p>
        </div>
      )}

      <div className="flex items-center justify-between pt-4 border-t border-gray-100">
        <div>
          <p className="text-xs text-gray-500 mb-0.5">{billingCycleLabels[server.billingCycle]}</p>
          <p className="text-lg font-semibold text-gray-900">₽{server.monthlyCost}</p>
        </div>
        <div className="text-right">
          <p className="text-xs text-gray-500 mb-0.5">Годовая стоимость</p>
          <p className="text-lg font-semibold text-indigo-600">₽{yearlyCost}</p>
        </div>
      </div>

      {showConfirmModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-sm w-full p-6">
            <h3 className="font-semibold text-gray-900 mb-4">Подтвердить оплату</h3>
            <p className="text-sm text-gray-600 mb-4">Сервер: {server.name}</p>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Дата оплаты</label>
                <input
                  type="date"
                  value={confirmDate}
                  onChange={e => setConfirmDate(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Сумма (₽)</label>
                <input
                  type="number"
                  step="0.01"
                  value={confirmAmount}
                  onChange={e => setConfirmAmount(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                />
              </div>
            </div>
            <div className="flex gap-2 mt-6">
              <button
                onClick={() => setShowConfirmModal(false)}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Отмена
              </button>
              <button
                disabled={saving}
                onClick={async () => {
                  setSaving(true);
                  await onConfirmPayment(server.id, confirmDate, parseFloat(confirmAmount) || undefined);
                  setSaving(false);
                  setShowConfirmModal(false);
                }}
                className="flex-1 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-70"
              >
                {saving ? 'Сохранение…' : 'Подтвердить'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
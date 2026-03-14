import { ServerStats } from '../types/server';
import { Calendar, DollarSign, AlertCircle } from 'lucide-react';

interface UpcomingPaymentsProps {
  stats: ServerStats;
}

export function UpcomingPayments({ stats }: UpcomingPaymentsProps) {
  if (stats.upcomingPayments.length === 0) {
    return (
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        <div className="flex items-center gap-2 mb-4">
          <Calendar className="w-5 h-5 text-indigo-600" />
          <h3 className="font-semibold text-gray-900">Предстоящие платежи (30 дней)</h3>
        </div>
        <div className="text-center py-8 text-gray-400">
          <AlertCircle className="w-12 h-12 mx-auto mb-3 opacity-50" />
          <p>Нет предстоящих платежей</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
      <div className="flex items-center gap-2 mb-4">
        <Calendar className="w-5 h-5 text-indigo-600" />
        <h3 className="font-semibold text-gray-900">Предстоящие платежи (30 дней)</h3>
      </div>
      
      <div className="space-y-3">
        {stats.upcomingPayments.map((payment) => {
          const isUrgent = payment.daysUntil <= 3;
          const isWarning = payment.daysUntil > 3 && payment.daysUntil <= 7;
          
          return (
            <div
              key={payment.serverId}
              className={`p-4 rounded-lg border transition-all ${
                isUrgent 
                  ? 'bg-red-50 border-red-200' 
                  : isWarning 
                  ? 'bg-orange-50 border-orange-200' 
                  : 'bg-gray-50 border-gray-200'
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <p className="font-medium text-gray-900 truncate">{payment.serverName}</p>
                    {isUrgent && (
                      <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 bg-red-100 text-red-700 rounded-full border border-red-200">
                        <AlertCircle className="w-3 h-3" />
                        Срочно
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-4 text-sm text-gray-600">
                    <span className="flex items-center gap-1.5">
                      <Calendar className="w-4 h-4" />
                      {new Date(payment.date).toLocaleDateString('ru-RU')}
                    </span>
                    <span className="flex items-center gap-1.5">
                      <DollarSign className="w-4 h-4" />
                      ₽{payment.amount}
                    </span>
                  </div>
                </div>
                <div className="text-right">
                  <p className={`text-2xl font-semibold ${
                    isUrgent 
                      ? 'text-red-600' 
                      : isWarning 
                      ? 'text-orange-600' 
                      : 'text-gray-900'
                  }`}>
                    {payment.daysUntil}
                  </p>
                  <p className="text-xs text-gray-500">
                    {payment.daysUntil === 1 ? 'день' : payment.daysUntil < 5 ? 'дня' : 'дней'}
                  </p>
                </div>
              </div>
            </div>
          );
        })}
      </div>
      
      {stats.upcomingPayments.length > 0 && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          <div className="flex items-center justify-between text-sm">
            <span className="text-gray-600">Общая сумма платежей:</span>
            <span className="font-semibold text-gray-900">
              ₽{stats.upcomingPayments.reduce((sum, p) => sum + p.amount, 0).toFixed(2)}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
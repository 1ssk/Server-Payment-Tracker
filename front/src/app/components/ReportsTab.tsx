import { useState, useEffect } from 'react';
import { Server, ReportRow } from '../types/server';
import { apiGetReports } from '../utils/api';
import { FileDown, Calendar, Filter } from 'lucide-react';

interface ReportsTabProps {
  servers: Server[];
}

export function ReportsTab({ servers }: ReportsTabProps) {
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const [serverId, setServerId] = useState('');
  const [rows, setRows] = useState<ReportRow[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const today = new Date();
    const y = today.getFullYear();
    const m = String(today.getMonth() + 1).padStart(2, '0');
    const d = String(today.getDate()).padStart(2, '0');
    if (!from) setFrom(`${y}-01-01`);
    if (!to) setTo(`${y}-${m}-${d}`);
  }, []);

  const load = async () => {
    setLoading(true);
    try {
      const data = await apiGetReports({ from: from || undefined, to: to || undefined, serverId: serverId || undefined });
      setRows(data);
    } catch (e) {
      console.error(e);
      setRows([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (from && to) load();
  }, [from, to, serverId]);

  const total = rows.reduce((s, r) => s + r.amount, 0);

  const exportCsv = () => {
    const header = 'Сервер;Дата оплаты;Сумма (₽)\n';
    const body = rows.map(r => `${r.serverName};${r.paidAt};${r.amount.toFixed(2)}`).join('\n');
    const blob = new Blob(['\uFEFF' + header + body], { type: 'text/csv;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `report-${from}-${to}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
      <div className="flex flex-wrap items-end gap-4 mb-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Период с</label>
          <input
            type="date"
            value={from}
            onChange={e => setFrom(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">по</label>
          <input
            type="date"
            value={to}
            onChange={e => setTo(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Сервер</label>
          <select
            value={serverId}
            onChange={e => setServerId(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg min-w-[200px]"
          >
            <option value="">Все серверы</option>
            {servers.map(s => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-70 flex items-center gap-2"
        >
          <Filter className="w-4 h-4" />
          Обновить
        </button>
        <button
          onClick={exportCsv}
          disabled={rows.length === 0}
          className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 flex items-center gap-2 disabled:opacity-50"
        >
          <FileDown className="w-4 h-4" />
          Выгрузить CSV
        </button>
      </div>
      <div className="overflow-x-auto">
        {loading ? (
          <p className="text-gray-500 py-8">Загрузка…</p>
        ) : rows.length === 0 ? (
          <p className="text-gray-500 py-8">Нет данных за выбранный период.</p>
        ) : (
          <>
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 text-left text-gray-600">
                  <th className="py-3 px-2">Сервер</th>
                  <th className="py-3 px-2">Дата оплаты</th>
                  <th className="py-3 px-2 text-right">Сумма (₽)</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r, i) => (
                  <tr key={i} className="border-b border-gray-100">
                    <td className="py-2 px-2">{r.serverName}</td>
                    <td className="py-2 px-2">{r.paidAt}</td>
                    <td className="py-2 px-2 text-right font-medium">₽{r.amount.toFixed(2)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div className="mt-4 pt-4 border-t border-gray-200 flex justify-end">
              <span className="font-semibold text-gray-900">Итого: ₽{total.toFixed(2)}</span>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

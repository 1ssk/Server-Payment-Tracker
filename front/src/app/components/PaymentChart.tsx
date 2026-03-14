import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { VPNServer } from '../types/server';
import { calculateYearlyCost } from '../utils/calculations';

interface PaymentChartProps {
  servers: VPNServer[];
}

export function PaymentChart({ servers }: PaymentChartProps) {
  const activeServers = servers.filter(s => s.status === 'active');
  
  const data = activeServers.map(server => ({
    name: server.name.length > 15 ? server.name.substring(0, 15) + '...' : server.name,
    fullName: server.name,
    cost: calculateYearlyCost(server)
  })).sort((a, b) => b.cost - a.cost);

  const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#3b82f6'];

  if (data.length === 0) {
    return (
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        <h3 className="font-semibold text-gray-900 mb-4">Годовые расходы по серверам</h3>
        <div className="h-64 flex items-center justify-center text-gray-400">
          <p>Нет активных серверов для отображения</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
      <h3 className="font-semibold text-gray-900 mb-4">Годовые расходы по серверам</h3>
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          <XAxis 
            dataKey="name" 
            tick={{ fill: '#6b7280', fontSize: 12 }}
            angle={-45}
            textAnchor="end"
            height={80}
          />
          <YAxis 
            tick={{ fill: '#6b7280', fontSize: 12 }}
            label={{ value: 'Стоимость (₽)', angle: -90, position: 'insideLeft', fill: '#6b7280' }}
          />
          <Tooltip 
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                return (
                  <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3">
                    <p className="font-medium text-gray-900">{payload[0].payload.fullName}</p>
                    <p className="text-indigo-600 font-semibold">₽{payload[0].value}/год</p>
                  </div>
                );
              }
              return null;
            }}
          />
          <Bar dataKey="cost" radius={[8, 8, 0, 0]}>
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
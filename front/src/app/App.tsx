import { useState, useEffect, useMemo } from 'react';
import { VPNServer } from './types/server';
import { calculateStats } from './utils/calculations';
import { apiCreateServer, apiDeleteServer, apiGetServers, apiUpdateServer } from './utils/api';
import { StatCard } from './components/StatCard';
import { ServerCard } from './components/ServerCard';
import { ServerForm } from './components/ServerForm';
import { PaymentChart } from './components/PaymentChart';
import { UpcomingPayments } from './components/UpcomingPayments';
import { LoginForm } from './components/LoginForm';
import { SettingsPanel } from './components/SettingsPanel';
import { 
  Server, 
  DollarSign, 
  TrendingUp, 
  Activity, 
  Plus, 
  Search,
  Filter,
  SortAsc,
  SortDesc,
  Settings,
  LogOut
} from 'lucide-react';

const AUTH_KEY = 'vpn-auth-token';

// Учетные данные администратора
const ADMIN_CREDENTIALS = {
  username: 'admin',
  password: 'admin123'
};

// Примеры серверов для демонстрации
const initialServers: VPNServer[] = [
  {
    id: '1',
    name: 'EU-Server-01',
    provider: 'DigitalOcean',
    providerUrl: 'https://digitalocean.com',
    ipAddress: '178.128.45.12',
    location: 'Нидерланды, Амстердам',
    monthlyCost: 1200.00,
    billingCycle: 'monthly',
    nextPaymentDate: '2026-03-20',
    status: 'active',
    notes: 'Основной сервер для европейских пользователей',
    createdAt: '2025-01-15'
  },
  {
    id: '2',
    name: 'US-Server-NYC',
    provider: 'Linode',
    providerUrl: 'https://linode.com',
    ipAddress: '45.79.132.89',
    location: 'США, Нью-Йорк',
    monthlyCost: 6000.00,
    billingCycle: 'quarterly',
    nextPaymentDate: '2026-04-01',
    status: 'active',
    notes: 'Высокопроизводительный сервер',
    createdAt: '2025-02-01'
  },
  {
    id: '3',
    name: 'ASIA-Server-SG',
    provider: 'Vultr',
    providerUrl: 'https://vultr.com',
    ipAddress: '103.253.145.78',
    location: 'Сингапур',
    monthlyCost: 15000.00,
    billingCycle: 'yearly',
    nextPaymentDate: '2026-12-15',
    status: 'active',
    createdAt: '2024-12-15'
  },
  {
    id: '4',
    name: 'EU-Backup-Server',
    provider: 'Hetzner',
    ipAddress: '95.217.28.156',
    location: 'Германия, Франкфурт',
    monthlyCost: 850.00,
    billingCycle: 'monthly',
    nextPaymentDate: '2026-03-18',
    status: 'inactive',
    notes: 'Резервный сервер, временно отключен',
    createdAt: '2025-03-01'
  }
];

type SortField = 'name' | 'location' | 'cost' | 'nextPayment';
type SortDirection = 'asc' | 'desc';

export default function App() {
  const [servers, setServers] = useState<VPNServer[]>([]);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<VPNServer | undefined>();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<VPNServer['status'] | 'all'>('all');
  const [sortField, setSortField] = useState<SortField>('name');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  // Проверка авторизации при загрузке
  useEffect(() => {
    const authToken = localStorage.getItem(AUTH_KEY);
    if (authToken === 'true') {
      setIsAuthenticated(true);
    }
  }, []);

  // Загрузка данных с backend
  useEffect(() => {
    const load = async () => {
      try {
        const data = await apiGetServers();
        if (data.length === 0) {
          setServers(initialServers);
        } else {
          setServers(data);
        }
      } catch (error) {
        console.error('Error loading servers from API:', error);
        setServers(initialServers);
      }
    };
    load();
  }, []);

  const stats = useMemo(() => calculateStats(servers), [servers]);

  const handleLogin = (username: string, password: string): boolean => {
    if (username === ADMIN_CREDENTIALS.username && password === ADMIN_CREDENTIALS.password) {
      localStorage.setItem(AUTH_KEY, 'true');
      setIsAuthenticated(true);
      return true;
    }
    return false;
  };

  const handleLogout = () => {
    localStorage.removeItem(AUTH_KEY);
    setIsAuthenticated(false);
  };

  const handleSaveServer = async (serverData: Omit<VPNServer, 'id' | 'createdAt'>) => {
    try {
      if (editingServer) {
        const updated: VPNServer = {
          ...editingServer,
          ...serverData,
        };
        const saved = await apiUpdateServer(updated);
        setServers(prev => prev.map(s => (s.id === saved.id ? saved : s)));
      } else {
        const created = await apiCreateServer({
          ...serverData,
        } as Omit<VPNServer, 'id' | 'createdAt'>);
        setServers(prev => [...prev, created]);
      }
      setIsFormOpen(false);
      setEditingServer(undefined);
    } catch (error) {
      console.error('Error saving server:', error);
      alert('Ошибка при сохранении сервера. Попробуйте ещё раз.');
    }
  };

  const handleEditServer = (server: VPNServer) => {
    setEditingServer(server);
    setIsFormOpen(true);
  };

  const handleDeleteServer = async (id: string) => {
    if (!confirm('Вы уверены, что хотите удалить этот сервер?')) {
      return;
    }
    try {
      await apiDeleteServer(id);
      setServers(prev => prev.filter(s => s.id !== id));
    } catch (error) {
      console.error('Error deleting server:', error);
      alert('Ошибка при удалении сервера. Попробуйте ещё раз.');
    }
  };

  const handleCancelForm = () => {
    setIsFormOpen(false);
    setEditingServer(undefined);
  };

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  // Фильтрация и сортировка
  const filteredAndSortedServers = useMemo(() => {
    let filtered = servers;

    // Поиск
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(server =>
        server.name.toLowerCase().includes(query) ||
        server.provider.toLowerCase().includes(query) ||
        server.location.toLowerCase().includes(query) ||
        server.ipAddress.includes(query)
      );
    }

    // Фильтр по статусу
    if (statusFilter !== 'all') {
      filtered = filtered.filter(server => server.status === statusFilter);
    }

    // Сортировка
    const sorted = [...filtered].sort((a, b) => {
      let compareValue = 0;

      switch (sortField) {
        case 'name':
          compareValue = a.name.localeCompare(b.name);
          break;
        case 'location':
          compareValue = a.location.localeCompare(b.location);
          break;
        case 'cost':
          compareValue = a.monthlyCost - b.monthlyCost;
          break;
        case 'nextPayment':
          compareValue = new Date(a.nextPaymentDate).getTime() - new Date(b.nextPaymentDate).getTime();
          break;
      }

      return sortDirection === 'asc' ? compareValue : -compareValue;
    });

    return sorted;
  }, [servers, searchQuery, statusFilter, sortField, sortDirection]);

  // Показываем форму логина если не авторизован
  if (!isAuthenticated) {
    return <LoginForm onLogin={handleLogin} />;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      {/* Панель управления пользователем */}
      <div className="fixed top-4 right-4 flex gap-3 z-40">
        <button
          onClick={() => setIsSettingsOpen(true)}
          className="px-4 py-2.5 bg-white border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors flex items-center gap-2 shadow-sm"
          title="Настройки SMTP"
        >
          <Settings className="w-5 h-5" />
          <span className="hidden sm:inline">Настройки</span>
        </button>
        <button
          onClick={handleLogout}
          className="px-4 py-2.5 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors flex items-center gap-2 shadow-sm"
          title="Выйти"
        >
          <LogOut className="w-5 h-5" />
          <span className="hidden sm:inline">Выйти</span>
        </button>
      </div>

      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Заголовок */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="p-3 bg-indigo-600 text-white rounded-xl shadow-lg">
              <Server className="w-8 h-8" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Управление VPN-серверами</h1>
              <p className="text-gray-600">Контроль серверов и расходов в одном месте</p>
            </div>
          </div>
        </div>

        {/* Статистика */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <StatCard
            title="Всего серверов"
            value={stats.totalServers}
            icon={Server}
            color="blue"
          />
          <StatCard
            title="Активных серверов"
            value={stats.activeServers}
            icon={Activity}
            color="green"
          />
          <StatCard
            title="Ежемесячно"
            value={`₽${stats.monthlyTotal.toFixed(2)}`}
            icon={DollarSign}
            color="orange"
          />
          <StatCard
            title="Годовая стоимость"
            value={`₽${stats.yearlyTotal.toFixed(2)}`}
            icon={TrendingUp}
            color="purple"
          />
        </div>

        {/* График и предстоящие платежи */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          <div className="lg:col-span-2">
            <PaymentChart servers={servers} />
          </div>
          <div>
            <UpcomingPayments stats={stats} />
          </div>
        </div>

        {/* Панель управления */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 mb-6">
          <div className="flex flex-col lg:flex-row gap-4 items-start lg:items-center justify-between mb-6">
            <div className="flex-1 w-full lg:w-auto">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  placeholder="Поиск по названию, провайдеру, локации или IP..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-colors"
                />
              </div>
            </div>
            <button
              onClick={() => setIsFormOpen(true)}
              className="w-full lg:w-auto px-6 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors flex items-center justify-center gap-2 shadow-sm"
            >
              <Plus className="w-5 h-5" />
              Добавить сервер
            </button>
          </div>

          <div className="flex flex-wrap gap-3 items-center">
            <div className="flex items-center gap-2">
              <Filter className="w-4 h-4 text-gray-500" />
              <span className="text-sm text-gray-600">Статус:</span>
            </div>
            <div className="flex gap-2">
              {(['all', 'active', 'inactive', 'pending'] as const).map((status) => (
                <button
                  key={status}
                  onClick={() => setStatusFilter(status)}
                  className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                    statusFilter === status
                      ? 'bg-indigo-600 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {status === 'all' ? 'Все' : status === 'active' ? 'Активные' : status === 'inactive' ? 'Неактивные' : 'Ожидание'}
                </button>
              ))}
            </div>

            <div className="w-px h-6 bg-gray-300 mx-2" />

            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-600">Сортировка:</span>
            </div>
            <div className="flex gap-2 flex-wrap">
              {[
                { field: 'name' as const, label: 'Название' },
                { field: 'location' as const, label: 'Локация' },
                { field: 'cost' as const, label: 'Стоимость' },
                { field: 'nextPayment' as const, label: 'Платёж' }
              ].map(({ field, label }) => (
                <button
                  key={field}
                  onClick={() => toggleSort(field)}
                  className={`px-3 py-1.5 rounded-lg text-sm transition-colors flex items-center gap-1.5 ${
                    sortField === field
                      ? 'bg-indigo-600 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {label}
                  {sortField === field && (
                    sortDirection === 'asc' ? <SortAsc className="w-3.5 h-3.5" /> : <SortDesc className="w-3.5 h-3.5" />
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Список серверов */}
        {filteredAndSortedServers.length === 0 ? (
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-12 text-center">
            <Server className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {searchQuery || statusFilter !== 'all' ? 'Серверы не найдены' : 'Нет серверов'}
            </h3>
            <p className="text-gray-500 mb-6">
              {searchQuery || statusFilter !== 'all' 
                ? 'Попробуйте изменить параметры поиска или фильтры'
                : 'Добавьте первый VPN-сервер, чтобы начать отслеживать расходы'
              }
            </p>
            {!searchQuery && statusFilter === 'all' && (
              <button
                onClick={() => setIsFormOpen(true)}
                className="px-6 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors inline-flex items-center gap-2"
              >
                <Plus className="w-5 h-5" />
                Добавить сервер
              </button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {filteredAndSortedServers.map(server => (
              <ServerCard
                key={server.id}
                server={server}
                onEdit={handleEditServer}
                onDelete={handleDeleteServer}
              />
            ))}
          </div>
        )}

        {/* Результаты поиска/фильтрации */}
        {(searchQuery || statusFilter !== 'all') && filteredAndSortedServers.length > 0 && (
          <div className="mt-4 text-center text-sm text-gray-600">
            Показано {filteredAndSortedServers.length} из {servers.length} серверов
          </div>
        )}
      </div>

      {/* Форма добавления/редактирования */}
      {isFormOpen && (
        <ServerForm
          server={editingServer}
          onSave={handleSaveServer}
          onCancel={handleCancelForm}
        />
      )}

      {/* Панель настроек */}
      {isSettingsOpen && (
        <SettingsPanel onClose={() => setIsSettingsOpen(false)} />
      )}
    </div>
  );
}

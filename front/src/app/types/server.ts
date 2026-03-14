export interface VPNServer {
  id: string;
  name: string;
  provider: string;
  providerUrl?: string;
  ipAddress: string;
  location: string;
  monthlyCost: number;
  billingCycle: 'monthly' | 'quarterly' | 'yearly';
  nextPaymentDate: string;
  status: 'active' | 'inactive' | 'pending';
  notes?: string;
  createdAt: string;
}

export interface ServerStats {
  totalServers: number;
  activeServers: number;
  monthlyTotal: number;
  yearlyTotal: number;
  upcomingPayments: Array<{
    serverId: string;
    serverName: string;
    amount: number;
    date: string;
    daysUntil: number;
  }>;
}

export interface SMTPSettings {
  host: string;
  port: number;
  username: string;
  password: string;
  from: string;
  to: string;
  enabled: boolean;
}
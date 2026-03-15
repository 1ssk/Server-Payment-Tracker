export interface Server {
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

/** @deprecated используйте Server */
export type VPNServer = Server;

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
  /** За сколько дней до срока присылать напоминание (по умолчанию 10). */
  reminderDaysBefore: number;
}

export interface Payment {
  id: string;
  serverId: string;
  amount: number;
  paidAt: string;
  createdAt: string;
}

export interface ReportRow {
  serverId: string;
  serverName: string;
  paidAt: string;
  amount: number;
}
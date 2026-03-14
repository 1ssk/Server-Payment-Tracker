import { VPNServer, ServerStats } from '../types/server';

export function calculateYearlyCost(server: VPNServer): number {
  switch (server.billingCycle) {
    case 'monthly':
      return server.monthlyCost * 12;
    case 'quarterly':
      return server.monthlyCost * 4;
    case 'yearly':
      return server.monthlyCost;
    default:
      return 0;
  }
}

export function getDaysUntilPayment(dateString: string): number {
  const paymentDate = new Date(dateString);
  const today = new Date();
  const diffTime = paymentDate.getTime() - today.getTime();
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
}

export function calculateStats(servers: VPNServer[]): ServerStats {
  const activeServers = servers.filter(s => s.status === 'active');
  
  const monthlyTotal = activeServers.reduce((sum, server) => {
    const monthlyCost = server.billingCycle === 'monthly' 
      ? server.monthlyCost 
      : server.billingCycle === 'quarterly'
      ? server.monthlyCost / 3
      : server.monthlyCost / 12;
    return sum + monthlyCost;
  }, 0);

  const yearlyTotal = activeServers.reduce((sum, server) => {
    return sum + calculateYearlyCost(server);
  }, 0);

  const upcomingPayments = servers
    .filter(s => s.status === 'active')
    .map(server => ({
      serverId: server.id,
      serverName: server.name,
      amount: server.monthlyCost,
      date: server.nextPaymentDate,
      daysUntil: getDaysUntilPayment(server.nextPaymentDate)
    }))
    .filter(payment => payment.daysUntil >= 0 && payment.daysUntil <= 30)
    .sort((a, b) => a.daysUntil - b.daysUntil);

  return {
    totalServers: servers.length,
    activeServers: activeServers.length,
    monthlyTotal,
    yearlyTotal,
    upcomingPayments
  };
}

export function getNextPaymentDate(currentDate: string, cycle: VPNServer['billingCycle']): string {
  const date = new Date(currentDate);
  
  switch (cycle) {
    case 'monthly':
      date.setMonth(date.getMonth() + 1);
      break;
    case 'quarterly':
      date.setMonth(date.getMonth() + 3);
      break;
    case 'yearly':
      date.setFullYear(date.getFullYear() + 1);
      break;
  }
  
  return date.toISOString().split('T')[0];
}

<template>
  <div class="account-dashboard">
    <header class="page-header">
      <h1>Личный кабинет</h1>
      <p class="welcome-message">Добро пожаловать, {{ getUserDisplayName() }}!</p>
    </header>

    <!-- User Information Card -->
    <div class="card mb-6">
      <div class="card-header">
        <h2 class="card-title">
          <span class="material-icons">person</span>
          Информация о пользователе
        </h2>
      </div>
      <div class="card-body">
        <div v-if="user" class="user-info-grid">
          <div class="info-item">
            <label>Пользователь:</label>
            <span class="value">{{ user.username }}</span>
          </div>
          <div class="info-item">
            <label>Роль:</label>
            <span class="badge role-badge" :class="getRoleClass(user.role)">
              {{ getRoleLabel(user.role) }}
            </span>
          </div>
          <div class="info-item">
            <label>Дата создания:</label>
            <span class="value">{{ formatDate(user.created_at) }}</span>
          </div>
        </div>
        <div v-else class="loading">
          <span class="material-icons">hourglass_empty</span>
          Загрузка данных пользователя...
        </div>
      </div>
    </div>

    <!-- Statistics Overview -->
    <div class="stats-grid mb-6">
      <!-- Contracts Stats -->
      <div class="card stat-card">
        <div class="stat-icon contracts-icon">
          <span class="material-icons">description</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-label">Договоры</h3>
          <p class="stat-value">{{ stats.contracts.total }}</p>
          <p class="stat-detail">{{ stats.contracts.active }} активных</p>
        </div>
      </div>

      <!-- Clients Stats -->
      <div class="card stat-card">
        <div class="stat-icon clients-icon">
          <span class="material-icons">people</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-label">Клиенты</h3>
          <p class="stat-value">{{ stats.clients.total }}</p>
          <p class="stat-detail">{{ stats.clients.new_this_month }} новых</p>
        </div>
      </div>

      <!-- Connections Stats -->
      <div class="card stat-card">
        <div class="stat-icon connections-icon">
          <span class="material-icons">router</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-label">Подключения</h3>
          <p class="stat-value">{{ stats.connections.total }}</p>
          <p class="stat-detail">{{ stats.connections.active }} активных</p>
        </div>
      </div>

      <!-- Revenue Stats -->
      <div class="card stat-card">
        <div class="stat-icon revenue-icon">
          <span class="material-icons">monetization_on</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-label">Выручка</h3>
          <p class="stat-value">{{ formatCurrency(stats.revenue.total) }}</p>
          <p class="stat-detail">за текущий месяц</p>
        </div>
      </div>
    </div>

    <!-- Traffic Statistics -->
    <div class="card mb-6">
      <div class="card-header">
        <h2 class="card-title">
          <span class="material-icons">analytics</span>
          Статистика трафика
        </h2>
      </div>
      <div class="card-body">
        <div class="traffic-stats-grid">
          <div class="traffic-stat-item">
            <div class="stat-icon-small incoming">
              <span class="material-icons">download</span>
            </div>
            <div>
              <h4>Входящий трафик</h4>
              <p class="traffic-value">{{ formatBytes(trafficStats.total_bytes_in) }}</p>
              <small>за последние 30 дней</small>
            </div>
          </div>
          
          <div class="traffic-stat-item">
            <div class="stat-icon-small outgoing">
              <span class="material-icons">upload</span>
            </div>
            <div>
              <h4>Исходящий трафик</h4>
              <p class="traffic-value">{{ formatBytes(trafficStats.total_bytes_out) }}</p>
              <small>за последние 30 дней</small>
            </div>
          </div>
          
          <div class="traffic-stat-item">
            <div class="stat-icon-small total">
              <span class="material-icons">swap_vert</span>
            </div>
            <div>
              <h4>Общий трафик</h4>
              <p class="traffic-value">{{ formatBytes(trafficStats.total_traffic) }}</p>
              <small>за последние 30 дней</small>
            </div>
          </div>
          
          <div class="traffic-stat-item">
            <div class="stat-icon-small active">
              <span class="material-icons">schedule</span>
            </div>
            <div>
              <h4>Активных дней</h4>
              <p class="traffic-value">{{ trafficStats.active_days }}</p>
              <small>с трафиком</small>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Payments -->
    <div class="card mb-6">
      <div class="card-header">
        <h2 class="card-title">
          <span class="material-icons">payment</span>
          Последние платежи
        </h2>
        <button class="btn btn-sm btn-primary">
          <span class="material-icons icon-sm">add</span>
          Добавить платеж
        </button>
      </div>
      <div class="card-body">
        <div v-if="recentPayments.length > 0" class="payments-list">
          <div v-for="payment in recentPayments" :key="payment.id" class="payment-item">
            <div class="payment-info">
              <h4>{{ payment.description }}</h4>
              <p class="payment-details">
                {{ payment.client_name }} • {{ formatDate(payment.date) }}
              </p>
            </div>
            <div class="payment-amount" :class="payment.type">
              {{ payment.type === 'income' ? '+' : '-' }}{{ formatCurrency(payment.amount) }}
            </div>
          </div>
        </div>
        <div v-else class="empty-state">
          <span class="material-icons">receipt</span>
          <p>Платежи отсутствуют</p>
          <small>Здесь будут отображаться последние транзакции</small>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions">
      <h3>Быстрые действия</h3>
      <div class="actions-grid">
        <router-link to="/clients" class="action-card">
          <span class="material-icons">person_add</span>
          <span>Добавить клиента</span>
        </router-link>
        
        <router-link to="/contracts" class="action-card">
          <span class="material-icons">note_add</span>
          <span>Создать договор</span>
        </router-link>
        
        <router-link to="/connections" class="action-card">
          <span class="material-icons">router</span>
          <span>Настроить подключение</span>
        </router-link>
        
        <router-link to="/traffic" class="action-card">
          <span class="material-icons">analytics</span>
          <span>Анализ трафика</span>
        </router-link>
      </div>
    </div>

    <!-- Logout Button -->
    <div class="logout-section">
      <button @click="logout" class="btn btn-outline logout-btn">
        <span class="material-icons">logout</span>
        Выйти из системы
      </button>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue';
import useAuth from '../composables/useAuth';

export default {
  name: 'AccountView',
  setup() {
    const { user, fetchUser, logout } = useAuth();
    
    // Mock data for statistics (to be replaced with real API calls)
    const stats = ref({
      contracts: {
        total: 1247,
        active: 1180
      },
      clients: {
        total: 892,
        new_this_month: 23
      },
      connections: {
        total: 1456,
        active: 1389
      },
      revenue: {
        total: 2847500 // в копейках
      }
    });

    const trafficStats = ref({
      total_bytes_in: 15680000000, // ~15.68 GB
      total_bytes_out: 4720000000, // ~4.72 GB
      total_traffic: 20400000000,  // ~20.4 GB
      active_days: 28
    });

    // Mock payment data
    const recentPayments = ref([
      {
        id: 1,
        description: 'Оплата за интернет',
        client_name: 'ООО "Рога и копыта"',
        date: '2024-01-15',
        amount: 25000, // в копейках
        type: 'income'
      },
      {
        id: 2,
        description: 'Возврат средств',
        client_name: 'Иванов И.И.',
        date: '2024-01-14',
        amount: 5000,
        type: 'expense'
      },
      {
        id: 3,
        description: 'Подключение тарифа',
        client_name: 'ЗАО "Технологии"',
        date: '2024-01-13',
        amount: 15000,
        type: 'income'
      }
    ]);

    const formatDate = (dateString) => {
      if (!dateString) return 'Не указано';
      const date = new Date(dateString);
      return date.toLocaleDateString('ru-RU');
    };

    const formatBytes = (bytes) => {
      if (bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatCurrency = (kopecks) => {
      const rubles = kopecks / 100;
      return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'RUB'
      }).format(rubles);
    };

    const getUserDisplayName = () => {
      if (!user.value) return '';
      return user.value.username;
    };

    const getRoleLabel = (role) => {
      const roleLabels = {
        admin: 'Администратор',
        manager: 'Менеджер',
        client: 'Клиент'
      };
      return roleLabels[role] || role;
    };

    const getRoleClass = (role) => {
      const roleClasses = {
        admin: 'badge-danger',
        manager: 'badge-warning',
        client: 'badge-primary'
      };
      return roleClasses[role] || 'badge-secondary';
    };

    onMounted(async () => {
      await fetchUser();
      // Here you would typically load real statistics from APIs
      // await loadStatistics();
      // await loadTrafficStats();
      // await loadRecentPayments();
    });

    return {
      user,
      logout,
      stats,
      trafficStats,
      recentPayments,
      formatDate,
      formatBytes,
      formatCurrency,
      getUserDisplayName,
      getRoleLabel,
      getRoleClass
    };
  }
};
</script>

<style scoped>
.account-dashboard {
  max-width: 100%;
  padding: 0;
}

.welcome-message {
  color: var(--gray-600);
  margin: 0;
}

.user-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-item label {
  font-weight: 500;
  color: var(--gray-700);
  font-size: 0.875rem;
}

.info-item .value {
  font-weight: 600;
  color: var(--gray-900);
}

.role-badge {
  max-width: fit-content;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.contracts-icon { background: linear-gradient(135deg, var(--primary-500), var(--primary-600)); }
.clients-icon { background: linear-gradient(135deg, var(--success-500), var(--success-600)); }
.connections-icon { background: linear-gradient(135deg, var(--warning-500), var(--warning-600)); }
.revenue-icon { background: linear-gradient(135deg, var(--error-500), var(--error-600)); }

.stat-content {
  flex: 1;
}

.stat-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--gray-600);
  margin: 0 0 0.25rem 0;
}

.stat-value {
  font-size: 1.875rem;
  font-weight: 700;
  color: var(--gray-900);
  margin: 0;
  line-height: 1;
}

.stat-detail {
  font-size: 0.75rem;
  color: var(--gray-500);
  margin: 0.25rem 0 0 0;
}

.traffic-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.traffic-stat-item {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon-small {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.incoming { background: var(--success-500); }
.outgoing { background: var(--primary-500); }
.total { background: var(--warning-500); }
.active { background: var(--error-500); }

.traffic-stat-item h4 {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--gray-700);
  margin: 0;
}

.traffic-value {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--gray-900);
  margin: 0.25rem 0;
}

.traffic-stat-item small {
  color: var(--gray-500);
  font-size: 0.75rem;
}

.payments-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.payment-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--gray-50);
  border-radius: 8px;
  border: 1px solid var(--gray-200);
}

.payment-info h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--gray-900);
  margin: 0 0 0.25rem 0;
}

.payment-details {
  font-size: 0.75rem;
  color: var(--gray-600);
  margin: 0;
}

.payment-amount {
  font-size: 1rem;
  font-weight: 600;
}

.payment-amount.income {
  color: var(--success-600);
}

.payment-amount.expense {
  color: var(--error-600);
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--gray-500);
}

.empty-state .material-icons {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.quick-actions {
  margin-bottom: 2rem;
}

.quick-actions h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--gray-900);
  margin: 0 0 1rem 0;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1.5rem 1rem;
  background: white;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  text-decoration: none;
  color: var(--gray-700);
  transition: all 0.2s ease;
}

.action-card:hover {
  color: var(--primary-600);
  border-color: var(--primary-300);
  background: var(--primary-50);
  transform: translateY(-2px);
}

.logout-section {
  text-align: center;
  padding: 2rem 0;
  border-top: 1px solid var(--gray-200);
}

.logout-btn {
  gap: 0.5rem;
}

.loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--gray-500);
  padding: 1rem;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .traffic-stats-grid {
    grid-template-columns: 1fr;
  }
  
  .actions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .payment-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
}
</style>
<script setup>
import { ref, computed } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import IssueForm from '@/components/forms/IssueForm.vue';
import IssueHistory from '@/components/IssueHistory.vue';
import StatusBadge from '@/components/ui/StatusBadge.vue';
import { formatDate, formatDateOptional } from '@/utils/dateUtils';
import apiClient from '@/api/client';
import { useNotificationStore } from '@/stores/notification';

const notificationStore = useNotificationStore();
const { items: issues, loading, createItem, updateItem, deleteItem } = useCrud('issues');

const isModalOpen = ref(false);
const currentIssue = ref(null);
const isEditMode = ref(false);
const statusFilter = ref('all');
const showHistory = ref(false);
const historyIssueId = ref(null);
const showUnresolveModal = ref(false);
const unresolveIssueId = ref(null);
const unresolveReason = ref('');

// Comment-related state
const expandedComments = ref(new Set());
const issueComments = ref({});
const loadingComments = ref(new Set());
const newCommentText = ref({});
const addingComment = ref(new Set());

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'title', label: 'Название' },
  { key: 'status', label: 'Статус', format: (value) => value === 'new' ? 'Новая' : 'Решена' },
  { key: 'created_at', label: 'Создана', format: (value) => formatDate(value) },
  { key: 'resolved_at', label: 'Решена', format: (value) => formatDateOptional(value) },
];

const filteredIssues = computed(() => {
  if (statusFilter.value === 'all') return issues.value;
  return issues.value.filter(issue => issue.status === statusFilter.value);
});

function openCreateModal() {
  isEditMode.value = false;
  currentIssue.value = { 
    title: '', 
    description: '',
    created_by: 1 // TODO: Get from auth context
  };
  isModalOpen.value = true;
}

function openEditModal(issue) {
  isEditMode.value = true;
  currentIssue.value = { ...issue };
  isModalOpen.value = true;
}

async function handleSave(issueData) {
  try {
    if (isEditMode.value) {
      await updateItem(issueData.id, issueData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Обращение обновлено',
        message: 'Данные обращения успешно обновлены'
      });
    } else {
      await createItem(issueData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Обращение создано',
        message: 'Новое обращение успешно создано'
      });
    }
    isModalOpen.value = false;
    currentIssue.value = null;
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка сохранения',
      message: 'Не удалось сохранить задачу'
    });
  }
}

async function handleDelete(issueId) {
  try {
    await deleteItem(issueId);
    notificationStore.addNotification({
      type: 'success',
      title: 'Обращение удалено',
      message: 'Обращение успешно удалено'
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка удаления',
      message: 'Не удалось удалить задачу'
    });
  }
}

async function resolveIssue(issue) {
  {
    try {
      const response = await fetch(`/api/issues/${issue.id}/resolve`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ resolved_by: 1 }) // TODO: Get from auth context
      });
      
      if (response.ok) {
        // Update the issue in the local list instead of reloading
        const index = issues.value.findIndex(i => i.id === issue.id);
        if (index !== -1) {
          issues.value[index].status = 'resolved';
          issues.value[index].resolved_at = new Date().toISOString();
        }
      } else {
        throw new Error('Failed to resolve issue');
      }
    } catch (error) {
      notificationStore.addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось решить задачу'
      });
    }
  }
}

function getStatusClass(status) {
  return status === 'new' ? 'status-new' : 'status-resolved';
}

function showIssueHistory(issue) {
  historyIssueId.value = issue.id;
  showHistory.value = true;
}

function closeHistory() {
  showHistory.value = false;
  historyIssueId.value = null;
}

function openUnresolveModal(issue) {
  unresolveIssueId.value = issue.id;
  unresolveReason.value = '';
  showUnresolveModal.value = true;
}

function closeUnresolveModal() {
  showUnresolveModal.value = false;
  unresolveIssueId.value = null;
  unresolveReason.value = '';
}

async function confirmUnresolve() {
  if (!unresolveReason.value.trim()) {
    notificationStore.addNotification({
      type: 'warning',
      title: 'Необходимо указать',
      message: 'Укажите причину возврата в работу'
    });
    return;
  }

  try {
    const response = await fetch(`/api/issues/${unresolveIssueId.value}/unresolve`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({ 
        unresolve_reason: unresolveReason.value.trim(),
        unresolve_by: 1 // TODO: Get from auth context
      })
    });
    
    if (response.ok) {
      // Update the issue in the local list
      const index = issues.value.findIndex(i => i.id === unresolveIssueId.value);
      if (index !== -1) {
        issues.value[index].status = 'new';
        issues.value[index].resolved_at = null;
      }
      closeUnresolveModal();
    } else {
      const errorText = await response.text();
      throw new Error(errorText || 'Failed to unresolve issue');
    }
  } catch (error) {
    console.error('Error unresolving issue:', error);
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка',
      message: 'Не удалось вернуть задачу в работу: ' + error.message
    });
  }
}

// Comment functions
async function toggleComments(issue) {
  if (expandedComments.value.has(issue.id)) {
    expandedComments.value.delete(issue.id);
  } else {
    expandedComments.value.add(issue.id);
    await loadComments(issue.id);
  }
}

async function loadComments(issueId) {
  if (loadingComments.value.has(issueId)) return;
  
  loadingComments.value.add(issueId);
  try {
    const response = await apiClient.get(`/issues/${issueId}/comments`);
    issueComments.value[issueId] = response.data || [];
  } catch (error) {
    console.error('Failed to load comments:', error);
    issueComments.value[issueId] = [];
  } finally {
    loadingComments.value.delete(issueId);
  }
}

async function addComment(issueId) {
  const message = newCommentText.value[issueId]?.trim();
  if (!message) return;
  
  addingComment.value.add(issueId);
  try {
    const response = await apiClient.post(`/issues/${issueId}/comments`, {
      message: message
    });
    
    // Add the new comment to the list
    if (!issueComments.value[issueId]) {
      issueComments.value[issueId] = [];
    }
    issueComments.value[issueId].push(response.data);
    
    // Clear the input
    newCommentText.value[issueId] = '';
  } catch (error) {
    console.error('Failed to add comment:', error);
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка',
      message: 'Не удалось добавить комментарий'
    });
  } finally {
    addingComment.value.delete(issueId);
  }
}

function formatCommentDate(dateString) {
  return new Date(dateString).toLocaleString('ru-RU');
}

function getAuthorBadgeClass(role) {
  switch (role) {
    case 'admin': return 'author-admin';
    case 'manager': return 'author-manager';
    case 'client': return 'author-client';
    default: return 'author-default';
  }
}

function getAuthorLabel(role) {
  switch (role) {
    case 'admin': return 'Администратор';
    case 'manager': return 'Менеджер';
    case 'client': return 'Клиент';
    default: return 'Пользователь';
  }
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление доработками</h1>
      <div class="page-actions">
        <select v-model="statusFilter" class="filter-select">
          <option value="all">Все задачи</option>
          <option value="new">Новые</option>
          <option value="resolved">Решенные</option>
        </select>
      </div>
    </header>

    <div class="issues-list">
      <div v-if="loading" class="loading">Загрузка...</div>
      <div v-else-if="filteredIssues.length === 0" class="empty-state">
        <p>{{ statusFilter === 'all' ? 'Задач пока нет' : 'Нет задач с выбранным статусом' }}</p>
      </div>
      <div v-else>
        <div 
          v-for="issue in filteredIssues" 
          :key="issue.id" 
          class="issue-card"
          :class="getStatusClass(issue.status)"
        >
          <div class="issue-header">
            <h3>{{ issue.title }}</h3>
            <div class="issue-meta">
              <span class="issue-id">ID: {{ issue.id }}</span>
              <StatusBadge type="issue_status" :value="issue.status" size="small" />
            </div>
          </div>
          <div class="issue-description">
            {{ issue.description }}
          </div>
          <div class="issue-footer">
            <div class="issue-dates">
              <span>Создана: {{ formatDate(issue.created_at) }}</span>
              <span v-if="issue.resolved_at">
                Решена: {{ formatDate(issue.resolved_at) }}
              </span>
            </div>
            <div class="issue-actions">
              <button 
                v-if="issue.status === 'new'" 
                class="btn btn-icon btn-sm btn-success"
                @click="resolveIssue(issue)"
                title="Решить"
              >
                <span class="material-icons icon-sm">check</span>
              </button>
              <button 
                v-if="issue.status === 'resolved'" 
                class="btn btn-icon btn-sm btn-warning"
                @click="openUnresolveModal(issue)"
                title="Вернуть в работу"
              >
                <span class="material-icons icon-sm">refresh</span>
              </button>
              <button 
                class="btn btn-icon btn-sm btn-info" 
                @click="showIssueHistory(issue)"
                title="История"
              >
                <span class="material-icons icon-sm">history</span>
              </button>
              <button 
                class="btn btn-icon btn-sm btn-primary" 
                @click="toggleComments(issue)"
                title="Комментарии"
              >
                <span class="material-icons icon-sm">comment</span>
              </button>
              <button 
                v-if="issue.status === 'new'"
                class="btn btn-icon btn-sm btn-secondary" 
                @click="openEditModal(issue)"
                title="Редактировать"
              >
                <span class="material-icons icon-sm">edit</span>
              </button>
              <button 
                class="btn btn-icon btn-sm btn-danger" 
                @click="handleDelete(issue.id)"
                title="Удалить"
              >
                <span class="material-icons icon-sm">delete</span>
              </button>
            </div>
          </div>
          
          <!-- Comments section -->
          <div v-if="expandedComments.has(issue.id)" class="comments-section">
            <div v-if="loadingComments.has(issue.id)" class="loading-comments">
              <span class="material-icons icon-sm">hourglass_empty</span>
              Загрузка комментариев...
            </div>
            <div v-else>
              <div class="comments-header">
                <h4>Комментарии ({{ issueComments[issue.id]?.length || 0 }})</h4>
              </div>
              
              <div class="comments-list">
                <div v-if="!issueComments[issue.id]?.length" class="no-comments">
                  Комментариев пока нет
                </div>
                <div 
                  v-for="comment in issueComments[issue.id]" 
                  :key="comment.id" 
                  class="comment"
                >
                  <div class="comment-header">
                    <span :class="['author-badge', getAuthorBadgeClass(comment.author_role)]">
                      {{ getAuthorLabel(comment.author_role) }}
                    </span>
                    <span class="comment-date">{{ formatCommentDate(comment.created_at) }}</span>
                  </div>
                  <div class="comment-message">{{ comment.message }}</div>
                </div>
              </div>
              
              <div class="add-comment">
                <div class="comment-input">
                  <textarea 
                    v-model="newCommentText[issue.id]" 
                    placeholder="Добавить комментарий..."
                    rows="3"
                    class="form-control"
                  ></textarea>
                </div>
                <div class="comment-actions">
                  <button 
                    type="button"
                    class="btn btn-sm btn-primary"
                    @click="addComment(issue.id)"
                    :disabled="!newCommentText[issue.id]?.trim() || addingComment.has(issue.id)"
                  >
                    <span v-if="addingComment.has(issue.id)" class="material-icons icon-sm">hourglass_empty</span>
                    <span v-else class="material-icons icon-sm">send</span>
                    {{ addingComment.has(issue.id) ? 'Добавление...' : 'Отправить' }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <button class="fab" @click="openCreateModal">
      <span class="material-icons icon-lg">add</span>
    </button>

    <Modal :is-open="isModalOpen" @close="isModalOpen = false">
      <template #header>
        <h2>{{ isEditMode ? 'Редактировать задачу' : 'Создать задачу' }}</h2>
      </template>
      <IssueForm
        v-if="currentIssue"
        :initial-data="currentIssue"
        :is-edit-mode="isEditMode"
        @save="handleSave"
        @cancel="isModalOpen = false"
      />
    </Modal>

    <!-- Модальное окно для возврата в работу -->
    <Modal :is-open="showUnresolveModal" @close="closeUnresolveModal">
      <template #header>
        <h2>Вернуть задачу в работу</h2>
      </template>
      <div class="unresolve-form">
        <div class="form-group">
          <label class="form-label">Причина возврата в работу:</label>
          <textarea 
            v-model="unresolveReason" 
            class="form-control" 
            rows="3" 
            placeholder="Укажите причину возврата задачи в работу..."
            required
          ></textarea>
        </div>
        <div class="form-actions">
          <button 
            type="button" 
            class="btn btn-md btn-secondary" 
            @click="closeUnresolveModal"
          >
            <span class="material-icons icon-sm">close</span>
            Отмена
          </button>
          <button 
            type="button" 
            class="btn btn-md btn-warning" 
            @click="confirmUnresolve"
            :disabled="!unresolveReason.trim()"
          >
            <span class="material-icons icon-sm">refresh</span>
            Вернуть в работу
          </button>
        </div>
      </div>
    </Modal>

    <!-- История изменений -->
    <IssueHistory
      v-if="historyIssueId"
      :issue-id="historyIssueId"
      :is-open="showHistory"
      @close="closeHistory"
    />
  </div>
</template>

<style scoped>
.btn-icon {
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.page-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.filter-select {
  padding: 0.5rem;
  border: 1px solid var(--gray-300);
  border-radius: 0.375rem;
  background: white;
}

.issues-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.issue-card {
  background: white;
  border: 1px solid var(--gray-200);
  border-radius: 0.5rem;
  padding: 1.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

.issue-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

.issue-card.status-resolved {
  background: var(--gray-50);
  border-color: var(--gray-300);
}

.issue-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.issue-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--gray-900);
}

.issue-meta {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.issue-id {
  font-size: 0.875rem;
  color: var(--gray-500);
}

.issue-status {
  padding: 0.25rem 0.75rem;
  border-radius: 1rem;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
}

.issue-status.status-new {
  background: var(--primary-100);
  color: var(--primary-700);
}

.issue-status.status-resolved {
  background: var(--green-100);
  color: var(--green-700);
}

.issue-description {
  color: var(--gray-700);
  line-height: 1.5;
  margin-bottom: 1rem;
  white-space: pre-wrap;
}

.issue-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 1rem;
  border-top: 1px solid var(--gray-200);
}

.issue-dates {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.875rem;
  color: var(--gray-500);
}

.issue-actions {
  display: flex;
  gap: 0.5rem;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: var(--gray-500);
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: var(--gray-500);
}

.empty-state p {
  margin: 0;
  font-size: 1.125rem;
}

@media (max-width: 768px) {
  .issue-header {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .issue-footer {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .issue-actions {
    justify-content: center;
  }
  
  .page-header {
    flex-direction: column;
    gap: 1rem;
  }
}

.unresolve-form {
  padding: 1rem;
}

.unresolve-form .form-group {
  margin-bottom: 1.5rem;
}

.unresolve-form .form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.unresolve-form textarea {
  min-height: 5rem;
  resize: vertical;
}

/* Comment styles */
.comments-section {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--gray-25);
  border-top: 1px solid var(--gray-200);
  border-radius: 0 0 0.5rem 0.5rem;
}

.loading-comments {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--gray-600);
  padding: 1rem;
  justify-content: center;
}

.comments-header h4 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--gray-900);
}

.comments-list {
  margin-bottom: 1rem;
}

.no-comments {
  text-align: center;
  color: var(--gray-500);
  padding: 1rem;
  font-style: italic;
}

.comment {
  background: white;
  border-radius: 0.5rem;
  padding: 1rem;
  margin-bottom: 0.75rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.author-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
}

.author-admin {
  background: var(--red-100);
  color: var(--red-700);
}

.author-manager {
  background: var(--blue-100);
  color: var(--blue-700);
}

.author-client {
  background: var(--green-100);
  color: var(--green-700);
}

.author-default {
  background: var(--gray-100);
  color: var(--gray-700);
}

.comment-date {
  font-size: 0.75rem;
  color: var(--gray-500);
}

.comment-message {
  color: var(--gray-800);
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.add-comment {
  border-top: 1px solid var(--gray-200);
  padding-top: 1rem;
}

.comment-input {
  margin-bottom: 0.75rem;
}

.comment-input textarea {
  width: 100%;
  border: 1px solid var(--gray-300);
  border-radius: 0.375rem;
  padding: 0.75rem;
  font-size: 0.875rem;
  line-height: 1.4;
  resize: vertical;
  min-height: 4rem;
}

.comment-input textarea:focus {
  outline: none;
  border-color: var(--primary-500);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.comment-actions {
  display: flex;
  justify-content: flex-end;
}

.comment-actions .btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
</style>
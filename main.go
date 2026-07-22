package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/webview/webview"
)

// ============ DATA STRUCTURES ============
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskManager struct {
	Tasks []Task `json:"tasks"`
}

type Stats struct {
	Total          int     `json:"total"`
	Completed      int     `json:"completed"`
	Pending        int     `json:"pending"`
	CompletionRate float64 `json:"completion_rate"`
}

var taskManager TaskManager
var dataFile = "tasks_data.json"

// ============ FILE OPERATIONS ============
func loadTasks() {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		taskManager.Tasks = []Task{}
		return
	}
	err = json.Unmarshal(data, &taskManager)
	if err != nil {
		taskManager.Tasks = []Task{}
	}
}

func saveTasks() {
	data, err := json.MarshalIndent(taskManager, "", "  ")
	if err != nil {
		log.Println("Error saving tasks:", err)
		return
	}
	err = os.WriteFile(dataFile, data, 0644)
	if err != nil {
		log.Println("Error writing tasks:", err)
	}
}

// ============ TASK OPERATIONS ============
func addTask(title, description, priority string) Task {
	task := Task{
		ID:          len(taskManager.Tasks) + 1,
		Title:       title,
		Description: description,
		Priority:    priority,
		Completed:   false,
		CreatedAt:   time.Now(),
	}
	taskManager.Tasks = append(taskManager.Tasks, task)
	saveTasks()
	return task
}

func deleteTask(id int) bool {
	newTasks := []Task{}
	for _, task := range taskManager.Tasks {
		if task.ID != id {
			newTasks = append(newTasks, task)
		}
	}
	taskManager.Tasks = newTasks
	saveTasks()
	return true
}

func toggleTask(id int) *Task {
	for i, task := range taskManager.Tasks {
		if task.ID == id {
			taskManager.Tasks[i].Completed = !task.Completed
			saveTasks()
			return &taskManager.Tasks[i]
		}
	}
	return nil
}

func clearCompleted() []Task {
	newTasks := []Task{}
	for _, task := range taskManager.Tasks {
		if !task.Completed {
			newTasks = append(newTasks, task)
		}
	}
	taskManager.Tasks = newTasks
	saveTasks()
	return taskManager.Tasks
}

func getStats() Stats {
	stats := Stats{
		Total: len(taskManager.Tasks),
	}
	for _, task := range taskManager.Tasks {
		if task.Completed {
			stats.Completed++
		}
	}
	stats.Pending = stats.Total - stats.Completed
	if stats.Total > 0 {
		stats.CompletionRate = float64(stats.Completed) / float64(stats.Total) * 100
	}
	return stats
}

// ============ HTTP HANDLERS ============
func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Desktop App</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
            background: #f5f7fa;
            min-height: 100vh;
            overflow: hidden;
        }
        
        .title-bar {
            height: 48px;
            background: rgba(255,255,255,0.95);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border-bottom: 1px solid rgba(0,0,0,0.06);
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 0 16px;
            -webkit-app-region: drag;
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            z-index: 50;
        }
        
        .title-bar-left {
            -webkit-app-region: drag;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .title-bar-icon { font-size: 18px; line-height: 1; }
        .title-bar-title { 
            font-size: 13px; 
            font-weight: 600; 
            color: #1a1a2e;
            letter-spacing: 0.3px;
        }
        .title-bar-title span { color: #667eea; }
        
        .title-bar-actions {
            display: flex;
            gap: 6px;
            -webkit-app-region: no-drag;
        }
        
        .window-btn {
            width: 32px;
            height: 32px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.2s;
            background: transparent;
            color: #64748b;
            font-size: 12px;
            -webkit-app-region: no-drag;
        }
        
        .window-btn:hover { background: rgba(0,0,0,0.05); }
        .window-btn.close:hover { background: #ef4444; color: white; }
        .window-btn svg { width: 14px; height: 14px; fill: currentColor; }
        
        .main-content {
            margin-top: 48px;
            height: calc(100vh - 48px);
            overflow-y: auto;
            padding: 24px;
            background: #f5f7fa;
        }
        
        .task-item {
            animation: slideIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
        }
        
        @keyframes slideIn {
            from { opacity: 0; transform: translateY(10px) scale(0.98); }
            to { opacity: 1; transform: translateY(0) scale(1); }
        }
        
        .main-content::-webkit-scrollbar { width: 4px; }
        .main-content::-webkit-scrollbar-track { background: transparent; }
        .main-content::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 10px; }
        .main-content::-webkit-scrollbar-thumb:hover { background: #94a3b8; }
        
        .btn {
            transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
        }
        .btn:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .btn:active { transform: scale(0.96); }
        
        .task-checkbox {
            width: 18px;
            height: 18px;
            cursor: pointer;
            accent-color: #667eea;
            transition: all 0.2s;
            -webkit-app-region: no-drag;
        }
        .task-checkbox:hover { transform: scale(1.1); }
        
        .priority-badge {
            font-size: 10px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.3px;
            padding: 2px 10px;
            border-radius: 20px;
        }
        .priority-high { background: #fef2f2; color: #dc2626; }
        .priority-medium { background: #fffbeb; color: #d97706; }
        .priority-low { background: #f0fdf4; color: #059669; }
        
        .toast-container {
            position: fixed;
            top: 64px;
            right: 20px;
            z-index: 9999;
            display: flex;
            flex-direction: column;
            gap: 8px;
            -webkit-app-region: no-drag;
            max-width: 360px;
        }
        
        .toast {
            padding: 12px 16px;
            border-radius: 12px;
            color: white;
            font-size: 13px;
            font-weight: 500;
            box-shadow: 0 8px 24px rgba(0,0,0,0.12);
            animation: slideInRight 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        @keyframes slideInRight {
            from { opacity: 0; transform: translateX(40px) scale(0.9); }
            to { opacity: 1; transform: translateX(0) scale(1); }
        }
        
        .toast-success { background: linear-gradient(135deg, #10b981 0%, #059669 100%); }
        .toast-error { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%); }
        .toast-warning { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); }
        .toast-info { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); }
        
        .modal-overlay {
            position: fixed;
            inset: 0;
            background: rgba(0,0,0,0.4);
            backdrop-filter: blur(8px);
            -webkit-backdrop-filter: blur(8px);
            z-index: 99999;
            display: none;
            justify-content: center;
            align-items: center;
            animation: fadeIn 0.2s ease;
            -webkit-app-region: no-drag;
        }
        .modal-overlay.active { display: flex; }
        
        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
        
        .modal-box {
            background: white;
            border-radius: 20px;
            padding: 32px;
            max-width: 420px;
            width: 92%;
            box-shadow: 0 20px 60px rgba(0,0,0,0.2);
            animation: modalPop 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
            -webkit-app-region: no-drag;
        }
        
        @keyframes modalPop {
            from { opacity: 0; transform: scale(0.9) translateY(20px); }
            to { opacity: 1; transform: scale(1) translateY(0); }
        }
        
        .modal-icon { font-size: 48px; text-align: center; display: block; margin-bottom: 12px; }
        .modal-title { font-size: 20px; font-weight: 700; text-align: center; color: #0f172a; margin-bottom: 8px; }
        .modal-text { font-size: 14px; color: #64748b; text-align: center; line-height: 1.6; margin-bottom: 24px; }
        .modal-actions { display: flex; gap: 10px; justify-content: center; }
        .modal-actions .btn { min-width: 100px; -webkit-app-region: no-drag; }
        
        .empty-state { text-align: center; padding: 60px 20px; }
        .empty-state .icon { font-size: 56px; display: block; margin-bottom: 16px; }
        .empty-state h3 { font-size: 18px; font-weight: 600; color: #0f172a; margin-bottom: 6px; }
        .empty-state p { font-size: 14px; color: #94a3b8; }
        
        @media (max-width: 640px) {
            .main-content { padding: 16px; }
            .title-bar-title { font-size: 12px; }
            .window-btn { width: 28px; height: 28px; }
            .modal-box { padding: 24px; margin: 16px; }
        }
    </style>
</head>
<body>
    <div class="title-bar">
        <div class="title-bar-left">
            <span class="title-bar-icon">⚡</span>
            <span class="title-bar-title">Go <span>Desktop</span></span>
        </div>
        <div class="title-bar-actions">
            <button class="window-btn minimize" onclick="minimizeWindow()">
                <svg viewBox="0 0 10 10"><rect x="0" y="4.5" width="10" height="1" rx="0.5"/></svg>
            </button>
            <button class="window-btn maximize" onclick="maximizeWindow()">
                <svg viewBox="0 0 10 10"><rect x="1.5" y="1.5" width="7" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/></svg>
            </button>
            <button class="window-btn close" onclick="closeWindow()">
                <svg viewBox="0 0 10 10"><path d="M2 2 L8 8 M8 2 L2 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            </button>
        </div>
    </div>
    
    <div class="main-content">
        <div class="max-w-3xl mx-auto">
            <div class="flex items-center justify-between mb-6">
                <div>
                    <h1 class="text-2xl font-bold text-slate-900">Tasks</h1>
                    <p class="text-sm text-slate-500">Manage your daily tasks</p>
                </div>
                <div class="flex items-center gap-2 text-sm" id="stats">
                    <span class="px-3 py-1 bg-white rounded-full shadow-sm border border-slate-100">
                        <span class="font-medium" id="totalCount">0</span> total
                    </span>
                    <span class="px-3 py-1 bg-white rounded-full shadow-sm border border-slate-100">
                        ✅ <span class="font-medium" id="completedCount">0</span>
                    </span>
                    <span class="px-3 py-1 bg-white rounded-full shadow-sm border border-slate-100">
                        ⏳ <span class="font-medium" id="pendingCount">0</span>
                    </span>
                </div>
            </div>
            
            <div class="bg-white rounded-xl shadow-sm border border-slate-100 p-4 mb-4">
                <div class="flex flex-col sm:flex-row gap-3">
                    <input type="text" id="taskTitle" placeholder="What needs to be done?" 
                           class="flex-1 px-4 py-2.5 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:border-transparent transition text-sm" />
                    <input type="text" id="taskDescription" placeholder="Add a note..." 
                           class="flex-1 px-4 py-2.5 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:border-transparent transition text-sm" />
                    <select id="taskPriority" 
                            class="px-4 py-2.5 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:border-transparent transition text-sm bg-white">
                        <option value="low">🟢 Low</option>
                        <option value="medium" selected>🟡 Medium</option>
                        <option value="high">🔴 High</option>
                    </select>
                    <button onclick="addTask()" 
                            class="btn px-6 py-2.5 bg-indigo-500 hover:bg-indigo-600 text-white rounded-lg font-medium text-sm whitespace-nowrap">
                        Add Task
                    </button>
                </div>
            </div>
            
            <div class="bg-white rounded-xl shadow-sm border border-slate-100 p-4" id="taskList">
                <div class="empty-state">
                    <span class="icon">✨</span>
                    <h3>No tasks yet</h3>
                    <p>Start by adding your first task above</p>
                </div>
            </div>
            
            <div class="flex items-center justify-between mt-4 text-sm text-slate-400">
                <span>⚡ v1.0.0</span>
                <div class="flex gap-2">
                    <button onclick="refreshTasks()" class="btn px-3 py-1.5 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 text-xs font-medium">
                        🔄 Refresh
                    </button>
                    <button onclick="clearCompleted()" class="btn px-3 py-1.5 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 text-xs font-medium">
                        🧹 Clear Done
                    </button>
                    <button onclick="exportData()" class="btn px-3 py-1.5 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-600 text-xs font-medium">
                        💾 Export
                    </button>
                </div>
            </div>
        </div>
    </div>
    
    <div class="modal-overlay" id="modalOverlay">
        <div class="modal-box">
            <span class="modal-icon" id="modalIcon">⚠️</span>
            <h2 class="modal-title" id="modalTitle">Are you sure?</h2>
            <p class="modal-text" id="modalText">This action cannot be undone.</p>
            <div class="modal-actions">
                <button class="btn px-6 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg font-medium text-sm" onclick="closeModal()">Cancel</button>
                <button class="btn px-6 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg font-medium text-sm" id="modalConfirmBtn" onclick="confirmModal()">Confirm</button>
            </div>
        </div>
    </div>
    
    <div class="toast-container" id="toastContainer"></div>
    
    <script>
        function minimizeWindow() { if (window.minimizeApp) window.minimizeApp(); }
        function maximizeWindow() { if (window.maximizeApp) window.maximizeApp(); }
        function closeWindow() {
            showModal('👋', 'Exit Application', 'Are you sure you want to close?', 'Yes, Close', 'danger',
                function() { if (window.terminateApp) window.terminateApp(); }
            );
        }
        
        let modalCallback = null;
        function showModal(icon, title, text, confirmText, type, callback) {
            document.getElementById('modalIcon').textContent = icon || '⚠️';
            document.getElementById('modalTitle').textContent = title || 'Are you sure?';
            document.getElementById('modalText').textContent = text || 'This action cannot be undone.';
            const confirmBtn = document.getElementById('modalConfirmBtn');
            confirmBtn.textContent = confirmText || 'Confirm';
            confirmBtn.className = 'btn px-6 py-2 text-white rounded-lg font-medium text-sm';
            if (type === 'danger') confirmBtn.classList.add('bg-red-500', 'hover:bg-red-600');
            else if (type === 'success') confirmBtn.classList.add('bg-green-500', 'hover:bg-green-600');
            else confirmBtn.classList.add('bg-indigo-500', 'hover:bg-indigo-600');
            modalCallback = callback;
            document.getElementById('modalOverlay').classList.add('active');
        }
        function closeModal() { document.getElementById('modalOverlay').classList.remove('active'); modalCallback = null; }
        function confirmModal() { if (modalCallback) modalCallback(); closeModal(); }
        document.getElementById('modalOverlay').addEventListener('click', function(e) { if (e.target === this) closeModal(); });
        document.addEventListener('keydown', function(e) { if (e.key === 'Escape') closeModal(); });
        
        function showToast(message, type = 'info', duration = 3000) {
            const container = document.getElementById('toastContainer');
            const toast = document.createElement('div');
            toast.className = 'toast toast-' + type;
            const icons = { success: '✅', error: '❌', warning: '⚠️', info: 'ℹ️' };
            toast.innerHTML = '<span>' + (icons[type] || 'ℹ️') + '</span><span>' + message + '</span>';
            container.appendChild(toast);
            setTimeout(() => {
                toast.style.opacity = '0';
                toast.style.transform = 'translateX(40px) scale(0.9)';
                setTimeout(() => toast.remove(), 300);
            }, duration);
        }
        
        async function addTask() {
            const title = document.getElementById('taskTitle').value.trim();
            const description = document.getElementById('taskDescription').value.trim();
            const priority = document.getElementById('taskPriority').value;
            if (!title) { showToast('Please enter a task title', 'warning'); return; }
            try {
                const response = await fetch('/api/add_task', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ title, description, priority })
                });
                if (response.ok) {
                    document.getElementById('taskTitle').value = '';
                    document.getElementById('taskDescription').value = '';
                    refreshTasks();
                    showToast('Task "' + title + '" added', 'success');
                }
            } catch (error) { showToast('Failed to add task', 'error'); }
        }
        
        async function toggleTask(id) {
            try { await fetch('/api/toggle_task/' + id, { method: 'POST' }); refreshTasks(); } 
            catch (error) { showToast('Failed to update task', 'error'); }
        }
        
        async function deleteTask(id, title) {
            showModal('🗑️', 'Delete Task', 'Delete "' + (title || 'this task') + '"?', 'Delete', 'danger', async function() {
                try { await fetch('/api/delete_task/' + id, { method: 'DELETE' }); refreshTasks(); showToast('Task deleted', 'success'); } 
                catch (error) { showToast('Failed to delete task', 'error'); }
            });
        }
        
        async function clearCompleted() {
            const stats = await fetch('/api/stats').then(r => r.json());
            if (stats.completed === 0) { showToast('No completed tasks to clear', 'info'); return; }
            showModal('🧹', 'Clear Completed', 'Delete ' + stats.completed + ' completed task(s)?', 'Clear All', 'danger', async function() {
                try { await fetch('/api/clear_completed', { method: 'POST' }); refreshTasks(); showToast('Cleared ' + stats.completed + ' tasks', 'success'); } 
                catch (error) { showToast('Failed to clear tasks', 'error'); }
            });
        }
        
        async function refreshTasks() {
            try {
                const [tasksRes, statsRes] = await Promise.all([fetch('/api/tasks'), fetch('/api/stats')]);
                const tasks = await tasksRes.json();
                const stats = await statsRes.json();
                renderTasks(tasks);
                updateStats(stats);
            } catch (error) { console.error('Error:', error); }
        }
        
        function renderTasks(tasks) {
            const list = document.getElementById('taskList');
            if (!tasks || tasks.length === 0) {
                list.innerHTML = '<div class="empty-state"><span class="icon">✨</span><h3>No tasks yet</h3><p>Start by adding your first task above</p></div>';
                return;
            }
            list.innerHTML = tasks.map(task => 
                '<div class="task-item flex items-center gap-3 py-3 border-b border-slate-100 last:border-0">' +
                    '<input type="checkbox" class="task-checkbox" ' + (task.completed ? 'checked' : '') + ' onchange="toggleTask(' + task.id + ')" />' +
                    '<div class="flex-1 min-w-0">' +
                        '<p class="text-sm ' + (task.completed ? 'line-through text-slate-400' : 'text-slate-700') + '">' + escapeHtml(task.title) + '</p>' +
                        (task.description ? '<p class="text-xs text-slate-400 truncate">' + escapeHtml(task.description) + '</p>' : '') +
                    '</div>' +
                    '<span class="priority-badge priority-' + task.priority + '">' + task.priority + '</span>' +
                    '<span class="text-xs text-slate-400 whitespace-nowrap">' + formatDate(task.created_at) + '</span>' +
                    '<button onclick="deleteTask(' + task.id + ', \'' + escapeHtml(task.title) + '\')" class="text-slate-400 hover:text-red-500 transition text-sm p-1">✕</button>' +
                '</div>'
            ).join('');
        }
        
        function updateStats(stats) {
            document.getElementById('totalCount').textContent = stats.total || 0;
            document.getElementById('completedCount').textContent = stats.completed || 0;
            document.getElementById('pendingCount').textContent = stats.pending || 0;
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        function formatDate(dateString) {
            try {
                const date = new Date(dateString);
                const now = new Date();
                const diff = Math.floor((now - date) / 1000 / 60);
                if (diff < 1) return 'now';
                if (diff < 60) return diff + 'm';
                if (diff < 1440) return Math.floor(diff / 60) + 'h';
                return date.toLocaleDateString();
            } catch { return ''; }
        }
        
        async function exportData() {
            try {
                const response = await fetch('/api/export');
                const data = await response.json();
                const json = JSON.stringify(data, null, 2);
                const blob = new Blob([json], {type: 'application/json'});
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = 'tasks_' + new Date().toISOString().slice(0,10) + '.json';
                a.click();
                URL.revokeObjectURL(url);
                showToast('Data exported', 'success');
            } catch (error) { showToast('Failed to export', 'error'); }
        }
        
        document.addEventListener('DOMContentLoaded', () => {
            document.getElementById('taskTitle').addEventListener('keypress', (e) => { if (e.key === 'Enter') addTask(); });
            refreshTasks();
            setInterval(refreshTasks, 30000);
            setTimeout(() => showToast('🚀 Welcome!', 'info'), 600);
        });
        
        window.minimizeWindow = minimizeWindow;
        window.maximizeWindow = maximizeWindow;
        window.closeWindow = closeWindow;
        window.addTask = addTask;
        window.toggleTask = toggleTask;
        window.deleteTask = deleteTask;
        window.clearCompleted = clearCompleted;
        window.refreshTasks = refreshTasks;
        window.exportData = exportData;
        window.showModal = showModal;
        window.closeModal = closeModal;
        window.confirmModal = confirmModal;
        window.showToast = showToast;
    </script>
</body>
</html>`
	fmt.Fprint(w, html)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskManager.Tasks)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getStats())
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	task := addTask(req.Title, req.Description, req.Priority)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func toggleTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var id int
	fmt.Sscanf(r.URL.Path, "/api/toggle_task/%d", &id)
	task := toggleTask(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var id int
	fmt.Sscanf(r.URL.Path, "/api/delete_task/%d", &id)
	deleteTask(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func clearCompletedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tasks := clearCompleted()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"app":         "Go Desktop App",
		"version":     "1.0.0",
		"export_date": time.Now().Format(time.RFC3339),
		"tasks":       taskManager.Tasks,
	})
}

func main() {
	loadTasks()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/add_task", addTaskHandler)
	http.HandleFunc("/api/toggle_task/", toggleTaskHandler)
	http.HandleFunc("/api/delete_task/", deleteTaskHandler)
	http.HandleFunc("/api/clear_completed", clearCompletedHandler)
	http.HandleFunc("/api/export", exportHandler)

	go func() {
		fmt.Println("Starting server on http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle("Go Desktop App")
	w.SetSize(900, 700, webview.HintNone)

	// Webview does not natively support Minimize and Maximize endpoints out of the box,
	// so dummy functions prevent JavaScript errors while Terminate runs correctly.
	w.Bind("minimizeApp", func() {})
	w.Bind("maximizeApp", func() {})
	w.Bind("terminateApp", func() {
		w.Terminate()
	})

	w.Navigate("http://localhost:8080")
	w.Run()
}
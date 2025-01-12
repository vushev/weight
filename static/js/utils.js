// Помощни функции
function togglePasswordVisibility(inputId) {
    const input = document.getElementById(inputId);
    const button = input.nextElementSibling;
    
    if (input.type === 'password') {
        input.type = 'text';
        button.textContent = '🔒';
    } else {
        input.type = 'password';
        button.textContent = '👁️';
    }
}

function formatDate(date) {
    return new Date(date).toLocaleDateString('bg-BG', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });
}

function formatDateTime(date) {
    return new Date(date).toLocaleString('bg-BG', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatNumber(number, decimals = 1) {
    return number ? number.toFixed(decimals) : '-';
}

function showError(message) {
    alert(message || 'Възникна грешка');
}

function showSuccess(message) {
    alert(message || 'Операцията е успешна');
}

function getAuthHeaders() {
    return {
        'Content-Type': 'application/json; charset=utf-8',
        'Authorization': localStorage.getItem('token')
    };
}

async function isAuthenticated() {
    console.log('isAuthenticated utils');
    return await authState.isAuthenticated;
}

function getCurrentUser() {
    const userJson = localStorage.getItem('user');
    return userJson ? JSON.parse(userJson) : null;
}

function refreshCurrentView() {
    const currentView = localStorage.getItem('currentView') || 'stats';
    
    switch (currentView) {
        case 'stats':
            showStats();
            break;
        case 'users':
            showUsers();
            break;
        case 'friends':
            showFriends();
            break;
        case 'challenges':
            showChallenges();
            break;
        case 'settings':
            showSettings();
            break;
    }
} 
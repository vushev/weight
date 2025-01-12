// Функция за проверка дали потребителят е логнат
function isAuthenticated(init = false) {
    if (init) {
        console.log('isAuthenticated init');
        return authState.check();
    }

    return authState.isAuthenticated;
}

// Функция за зареждане на компоненти
async function loadComponent(name) {
    console.log('loadComponent', name);
    try {
        const response = await fetch(`/static/components/${name}.html`);
        const html = await response.text();
        document.getElementById('components').innerHTML = html;
    } catch (error) {
        console.error('Error loading component:', error);
    }
}

// Функция за инициализация на приложението
async function initializeApp() {
    // Скриваме всички секции първоначално
    hideAllSections();
    
    // Проверяваме за автентикация
    if (!isAuthenticated()) {
        // Ако потребителят е логнат, показваме навигацията
        // document.getElementById('mainNav').classList.remove('hidden');
        // Зареждаме началната страница
        await showWeight();
    } else {
        // Ако потребителят не е логнат, показваме формата за вход
        await loadComponent('auth');
        showLoginForm();
    }
}

// Функция за скриване на всички секции
function hideAllSections() {
    const sections = [
        'weightSection',
        'caloriesSection',
        'usersSection',
        'friendsSection',
        'challengesSection',
        'settingsSection',
    ];
    
    sections.forEach(sectionId => {
        const section = document.getElementById(sectionId);
        if (section) {
            section.style.display = 'none';
        }
    });

    const main = document.getElementsByTagName('main')[0];
    if (main) {
        const sections = main.querySelectorAll('div.section');
        sections.forEach(section => {
            section.classList.add('hidden');
        });
    }

    // document.getElementById('mainNav').classList.add('hidden');
}

// Функция за показване на секция
function showSection(sectionId) {
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }

    // Скриваме всички секции
    hideAllSections();

    // Показваме избраната секция
    const section = document.getElementById(sectionId);
    console.log('showSection', section);
    if (section) {
        section.classList.remove('hidden');
        section.style.display = 'block';
    }
}

// Функции за навигация
async function showWeight() {

    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }

    // showSection('weightSection');
    loadComponent('weight');
    try {
        await loadStats();
    } catch (error) {
        console.error('Error loading weight stats:', error);
        alert('Грешка при зареждане на статистиката за теглото');
    }
}

async function showCalories() {
    console.log('showCalories', isAuthenticated());
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }
    // showSection('caloriesSection');
    loadComponent('calories');
    try {
        await Promise.all([
            loadCalorieSettings(),
            loadCalorieStats(),
            loadCalorieIntake()
        ]);
    } catch (error) {
        console.error('Error loading calories data:', error);
        alert('Грешка при зареждане на данните за калориите');
    }
}

async function showUsers() {
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }
    // showSection('usersSection');
    loadComponent('users');
    try {
        await loadUsers();
    } catch (error) {
        console.error('Error loading users:', error);
        alert('Грешка при зареждане на потребителите');
    }
}

async function showFriends() {
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }
    // showSection('friendsSection');
    loadComponent('friends');
    try {
        await loadFriends();
    } catch (error) {
        console.error('Error loading friends:', error);
        alert('Грешка при зареждане на приятелите');
    }
}


async function showUserSettings() {
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }
    // showSection('userSettingsSection');
    loadComponent('settings');
    try {
        await loadUserSettings();
    } catch (error) {
        console.error('Error loading settings:', error);
        alert('Грешка при зареждане на настройките');
    }
}

// Инициализираме приложението при зареждане
window.onload = initializeApp; 
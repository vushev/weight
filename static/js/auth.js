// Функции за автентикация
async function register() {
    const username = document.getElementById('registerUsername').value;
    const password = document.getElementById('registerPassword').value;
    const height = parseFloat(document.getElementById('registerHeight').value);

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.register}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password, height })
        });

        if (response.ok) {
            alert('Регистрацията е успешна! Моля, влезте в системата.');
            loadComponent('auth');
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при регистрация');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Възникна грешка при комуникацията със сървъра');
    }
}

async function login() {
    console.log('Login function called');
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.login}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });

        if (response.ok) {
            const data = await response.json();
            localStorage.setItem('token', data.token);
            localStorage.setItem('user', JSON.stringify(data.user));
            document.getElementById('mainNav').style.display = 'flex';
            showStats();
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при вход');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Възникна грешка при комуникацията със сървъра');
    }
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    document.getElementById('mainNav').style.display = 'none';
    loadComponent('auth');
}

async function resetPassword() {
    const username = document.getElementById('resetUsername').value;
    
    if (!username) {
        alert('Моля, въведете потребителско име');
        return;
    }

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.resetPassword}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username })
        });

        if (response.ok) {
            const data = await response.json();
            alert(`Вашата нова парола е: ${data.newPassword}\nМоля, запомнете я и я променете след вход в системата.`);
            loadComponent('auth');
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при ресет на паролата');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Възникна грешка при комуникацията със сървъра');
    }
}

// Помощни функции
function showLogin() {
    document.getElementById('registerForm').style.display = 'none';
    document.getElementById('loginForm').style.display = 'block';
    document.getElementById('resetPasswordForm').style.display = 'none';
}

function showRegister() {
    document.getElementById('registerForm').style.display = 'block';
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('resetPasswordForm').style.display = 'none';
}

function showResetPassword() {
    document.getElementById('registerForm').style.display = 'none';
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('resetPasswordForm').style.display = 'block';
}

// Добавяме функция за показване на формата за вход
function showLoginForm() {
    document.getElementById('loginForm').style.display = 'block';
    document.getElementById('registerForm').style.display = 'none';
}

// Добавяме функция за инициализация
async function initAuth() {
    await loadComponent('auth');
    showLoginForm(); // Показваме формата за вход по подразбиране
}

// Обновяваме функцията за показване на регистрационната форма
function showRegisterForm() {
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('registerForm').style.display = 'block';
} 
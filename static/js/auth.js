// Функции за автентикация
async function register() {
    const username = document.getElementById('registerUsername').value;
    const password = document.getElementById('registerPassword').value;
    const height = parseFloat(document.getElementById('registerHeight').value);

    if (!username || !password || !height) {
        alert('Моля, въведете всички полета');
        return;
    }

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.register}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;',
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

    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;

    if (!username || !password) {
        alert('Моля, въведете потребителско име и парола');
        return;
    }

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.login}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json; charset=utf-8',
            },
            body: JSON.stringify({ username, password })
        });

        if (response.ok) {
            console.log('Response:', response);
            const data = await response.json();
            await authState.check();
            // localStorage.setItem('token', data.token);
            // localStorage.setItem('user', JSON.stringify(data.user));

            // Показваме навигацията и съдържанието
            // document.getElementById('mainNav').classList.remove('hidden');
            document.getElementById('navToggle').classList.remove('hidden');
            document.getElementById('mainNavAccordion').classList.remove('hidden');
            // document.getElementById('content').classList.remove('hidden');
            // showWeight();
            // document.getElementById('mainNav').classList.remove('hidden');

            document.getElementById('components').innerHTML = `<div class="container">Welcome ${authState.user.username}</div>`;

        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при вход');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Възникна грешка при комуникацията със сървъра');
    }
}

async function logout() {
    // localStorage.removeItem('token');
    // localStorage.removeItem('user');
    
    await api.logout();

    // Скриваме навигацията и съдържанието
    const main = document.getElementsByTagName('main')[0];
    if (main) {
        const sections = main.querySelectorAll('div.section');
        sections.forEach(section => {
            section.classList.add('hidden');
        });
    } else {
        console.log('No main element found');
    }
    
    document.getElementById('mainNav').classList.add('hidden');
    // document.getElementById('content').classList.add('hidden');
    // document.getElementById('components').classList.remove('hidden');
    
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
                'Content-Type': 'application/json; charset=utf-8',
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
    // const token = localStorage.getItem('token');
    const token = document.cookie.split('; ').find(row => row.startsWith('access_token=')).split('=')[1];
    console.log('Token:', token);
    if (token) {
        // Ако има токен, показваме съдържанието
        document.getElementById('mainNav').classList.remove('hidden');
        document.getElementById('content').classList.remove('hidden');
        document.getElementById('components').classList.add('hidden');
        // showStats();
    } else {
        // Ако няма токен, показваме формата за вход
        document.getElementById('mainNav').classList.add('hidden');
        document.getElementById('content').classList.add('hidden');
        document.getElementById('components').classList.remove('hidden');
        await loadComponent('auth');
        showLoginForm();
    }
}

// Обновяваме функцията за показване на регистрационната форма
function showRegisterForm() {
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('registerForm').style.display = 'block';
} 

const authState = {
    isAuthenticated: false,
    user: null,

    async check() {
        try {
            /*const response = await fetch(`${config.apiUrl}${config.endpoints.authStatus}`, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                    'Cache-Control': 'no-cache',
                    'Pragma': 'no-cache'
                },
                mode: 'same-origin'  // важно за same-origin заявки
            });*/

            const response = await api.checkAuth();

            if (response.ok !== true) {
                this.isAuthenticated = false;
                this.user = null;
                return false;
            }

            // const data = await response.json();
            const data = response;

            this.isAuthenticated = true;
            this.user = data.user;

            return true;
        } catch {
            this.isAuthenticated = false;
            this.user = null;
            return false;
        }
    }
};
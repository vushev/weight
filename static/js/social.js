// Глобални променливи за съхранение на данните
let allUsers = [];
let allFriends = [];

// Функции за търсене
function searchUsers(event) {
    const searchTerm = event.target.value.toLowerCase();
    
    const filteredUsers = allUsers.filter(user => 
        user.username.toLowerCase().includes(searchTerm) ||
        (user.firstName && user.firstName.toLowerCase().includes(searchTerm)) ||
        (user.lastName && user.lastName.toLowerCase().includes(searchTerm))
    );

    updateUsersList(filteredUsers);
}

function searchFriends(event) {
    const searchTerm = event.target.value.toLowerCase();
    
    const filteredFriends = allFriends.filter(friend => 
        friend.username.toLowerCase().includes(searchTerm) ||
        (friend.firstName && friend.firstName.toLowerCase().includes(searchTerm)) ||
        (user.lastName && user.lastName.toLowerCase().includes(searchTerm))
    );

    updateFriendsList(filteredFriends);
}

// Функции за зареждане на данни
async function loadUsers() {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.users}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при зареждане на потребителите');
        }

        const data = await response.json();
        allUsers = data || [];
        updateUsersList(allUsers);
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function loadFriends() {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.friends}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при зареждане на приятели');
        }

        const data = await response.json();
        allFriends = data || [];
        updateFriendsList(allFriends);
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

// Функции за обновяване на интерфейса
function updateUsersList(users) {
    const usersList = document.getElementById('usersList');
    if (!usersList) {
        console.error('Не е намерен елемент с id "usersList"');
        return;
    }
    
    usersList.innerHTML = '';
    
    if (!Array.isArray(users) || users.length === 0) {
        usersList.innerHTML = '<p>Няма намерени потребители</p>';
        return;
    }
    
    const currentUser = JSON.parse(localStorage.getItem('user'));
    
    users.forEach(user => {
        // Не показваме текущия потребител в списъка
        if (user.id === currentUser.id) return;
        
        usersList.innerHTML += `
            <div class="user-card">
                <div>
                    <strong>${user.username}</strong>
                    ${user.firstName || user.lastName ? 
                        `<p>${user.firstName || ''} ${user.lastName || ''}</p>` : ''}
                    <p>Прогрес: ${user.progress ? user.progress.toFixed(2) : 0}%</p>
                </div>
                <button onclick="sendFriendRequest('${user.id}')">
                    Добави приятел
                </button>
            </div>
        `;
    });
}

// UI компоненти
const UI = {
    createFriendCard(friend, currentUser) {
        const card = document.createElement('div');
        card.className = 'friend-card';
        
        const info = this.createFriendInfo(friend, currentUser);
        const buttons = this.createFriendButtons(friend, currentUser);
        
        card.appendChild(info);
        if (buttons) card.appendChild(buttons);
        
        return card;
    },

    createFriendInfo(friend, currentUser) {
        const info = document.createElement('div');
        
        const name = document.createElement('strong');
        name.textContent = friend.username || 'Неизвестен потребител';
        
        const status = document.createElement('p');
        status.textContent = `Статус: ${translateStatus(friend.status || 'pending')}`;
        
        const progress = document.createElement('p');
        progress.textContent = `Прогрес: ${friend.progress ? friend.progress.toFixed(2) : 0}%`;
        
        info.appendChild(name);
        info.appendChild(status);
        
        if (friend.status === 'pending') {
            const requestInfo = document.createElement('p');
            requestInfo.textContent = friend.senderId === currentUser.id ? 
                'Изпратена заявка' : 'Получена заявка';
            info.appendChild(requestInfo);
        }
        
        info.appendChild(progress);
        return info;
    },

    createFriendButtons(friend, currentUser) {
        if (!friend.status) return null;

        const buttonGroup = document.createElement('div');
        buttonGroup.className = 'button-group';

        if (friend.status === 'pending') {
            const friendshipId = friend.friendshipId || friend.id;

            const acceptBtn = document.createElement('button');
            acceptBtn.className = 'accept-btn';
            acceptBtn.textContent = 'Приеми';
            acceptBtn.onclick = () => acceptFriendRequest(friendshipId);

            const rejectBtn = document.createElement('button');
            rejectBtn.className = 'reject-btn';
            rejectBtn.textContent = 'Отхвърли';
            rejectBtn.onclick = () => rejectFriendRequest(friendshipId);

            buttonGroup.appendChild(acceptBtn);
            buttonGroup.appendChild(rejectBtn);
        }

        if (friend.status === 'accepted') {
            const challengeBtn = document.createElement('button');
            challengeBtn.textContent = 'Предизвикай';
            challengeBtn.onclick = () => window.showNewChallengeForm(friend.id);
            buttonGroup.appendChild(challengeBtn);
        }

        return buttonGroup.children.length > 0 ? buttonGroup : null;
    }
};

// Основна функция за обновяване на списъка
function updateFriendsList(friends) {
    const friendsList = document.getElementById('friendsList');
    if (!friendsList) {
        console.error('Не е намерен елемент с id "friendsList"');
        return;
    }
    
    friendsList.innerHTML = '';
    
    if (!Array.isArray(friends) || friends.length === 0) {
        const message = document.createElement('p');
        message.textContent = 'Все още нямате приятели';
        friendsList.appendChild(message);
        return;
    }

    const currentUser = JSON.parse(localStorage.getItem('user'));
    
    friends.forEach(friend => {
        if (!friend) return;
        const card = UI.createFriendCard(friend, currentUser);
        friendsList.appendChild(card);
    });
}

// Функции за управление на приятелства
async function sendFriendRequest(userId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.friendRequest}/${userId}`, {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при изпращане на заявката');
        }

        alert('Заявката е изпратена успешно');
        await loadUsers(); // Презареждаме списъка с потребители
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function acceptFriendRequest(friendshipId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.acceptFriend}/${friendshipId}`, {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при приемане на заявката');
        }

        await loadFriends(); // Презареждаме списъка с приятели
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function rejectFriendRequest(friendshipId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.rejectFriend}/${friendshipId}`, {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при отхвърляне на заявката');
        }

        await loadFriends(); // Презареждаме списъка с приятели
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

// Помощни функции
function translateStatus(status) {
    switch (status) {
        case 'pending': return 'Изчакваща';
        case 'accepted': return 'Приета';
        case 'rejected': return 'Отхвърлена';
        default: return status;
    }
}
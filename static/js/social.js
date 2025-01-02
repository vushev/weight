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
        (friend.lastName && friend.lastName.toLowerCase().includes(searchTerm))
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
        console.log('Заредени потребители:', allUsers); // Debug log
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
        console.log('Заредени приятели:', allFriends); // Debug log
        updateFriendsList(allFriends);
        updateOpponentSelect(allFriends);
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function loadChallenges() {
    try {
        console.log('Зареждане на предизвикателства...'); // Debug log
        const response = await fetch(`${config.apiUrl}${config.endpoints.challenges}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при зареждане на предизвикателствата');
        }

        const data = await response.json();
        console.log('Получени предизвикателства:', data); // Debug log
        
        if (!Array.isArray(data)) {
            console.error('Невалиден формат на данните:', data);
            throw new Error('Невалиден формат на данните от сървъра');
        }

        updateChallengesList(data);
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
    
    console.log('Обновяване на списъка с потребители:', users); // Debug log
    
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
                <button onclick="sendFriendRequest(${user.id})">
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

        console.log('Creating buttons for friend:', friend);

        if (friend.status === 'pending') {
            // Проверяваме дали имаме правилното ID
            const friendshipId = friend.friendshipId || friend.id;
            console.log('Using friendship ID:', friendshipId);

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
            challengeBtn.onclick = () => showNewChallengeForm(friend.id);
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

function updateOpponentSelect(friends) {
    const opponentSelect = document.getElementById('challengeOpponent');
    if (!opponentSelect) return;
    
    opponentSelect.innerHTML = '<option value="">Избери приятел</option>';
    friends.filter(friend => friend.status === 'accepted')
          .forEach(friend => {
              opponentSelect.innerHTML += `
                  <option value="${friend.id}">${friend.username}</option>
              `;
          });
}

function updateChallengesList(challenges) {
    const challengesList = document.getElementById('challengesList');
    if (!challengesList) {
        console.error('Не е намерен елемент с id "challengesList"');
        return;
    }
    
    console.log('Обновяване на списъка с предизвикателства:', challenges); // Debug log
    
    challengesList.innerHTML = '';
    
    if (!Array.isArray(challenges) || challenges.length === 0) {
        challengesList.innerHTML = '<p>Все още нямате предизвикателства</p>';
        return;
    }
    
    const currentUser = JSON.parse(localStorage.getItem('user'));
    
    challenges.forEach(challenge => {
        if (!challenge) return; // Пропускаме невалидни записи
        
        const startDate = new Date(challenge.startDate).toLocaleDateString('bg-BG');
        const endDate = new Date(challenge.endDate).toLocaleDateString('bg-BG');
        const isOpponent = challenge.opponentId === currentUser.id;
        const statusText = translateChallengeStatus(challenge.status);
        
        challengesList.innerHTML += `
            <div class="challenge-card">
                <div>
                    <strong>Предизвикателство</strong>
                    <p>От: ${startDate} До: ${endDate}</p>
                    <p>Статус: ${statusText}</p>
                    <p>Създател: ${challenge.creatorName || 'Неизвестен'}</p>
                    <p>Опонент: ${challenge.opponentName || 'Неизвестен'}</p>
                </div>
                ${challenge.status === 'pending' && isOpponent ? `
                    <div class="button-group">
                        <button onclick="acceptChallenge(${challenge.id})">Приеми</button>
                        <button onclick="rejectChallenge(${challenge.id})">Отхвърли</button>
                    </div>
                ` : ''}
                ${challenge.status === 'active' ? `
                    <button onclick="viewChallengeResults(${challenge.id})">Виж резултати</button>
                ` : ''}
            </div>
        `;
    });
}

function translateChallengeStatus(status) {
    const translations = {
        'pending': 'Изчакващо',
        'active': 'Активно',
        'completed': 'Завършено',
        'rejected': 'Отхвърлено'
    };
    return translations[status] || status;
}

// Функции за действия
async function sendFriendRequest(userId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.friendRequest.replace(':userId', userId)}`, {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при изпращане на заявката');
        }

        alert('Заявката е изпратена успешно');
        await loadUsers();
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function acceptFriendRequest(friendshipId) {
    try {
        console.log('Accepting friend request with ID:', friendshipId);
        
        // Проверяваме дали имаме ID
        if (!friendshipId) {
            console.error('No friendship ID provided');
            return;
        }

        const url = `${config.apiUrl}${config.endpoints.friendAccept.replace(':friendshipId', friendshipId)}`;
        console.log('Request URL:', url);

        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token'),
                'Content-Type': 'application/json'
            }
        });

        console.log('Response status:', response.status);

        if (response.ok) {
            alert('Приятелството е прието успешно!');
            await loadFriends();
        } else {
            const data = await response.json();
            console.error('Error data:', data);
            alert(data.error || 'Грешка при приемане на приятелството');
        }
    } catch (error) {
        console.error('Error accepting friend request:', error);
        alert('Грешка при комуникацията със сървъра');
    }
}

async function rejectFriendRequest(friendshipId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.friendReject.replace(':friendshipId', friendshipId)}`, {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (response.ok) {
            alert('Заявката е отхвърлена');
            // Презареждаме списъка с приятели
            loadFriends();
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при отхвърляне на заявката');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Грешка при комуникацията със сървъра');
    }
}

async function acceptChallenge(challengeId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.challengeAccept.replace(':challengeId', challengeId)}`, {
            method: 'PUT',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при приемане на предизвикателството');
        }

        await loadChallenges();
        alert('Предизвикателството е прието успешно');
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function rejectChallenge(challengeId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.challengeReject.replace(':challengeId', challengeId)}`, {
            method: 'PUT',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при отхвърляне на предизвикателството');
        }

        await loadChallenges();
        alert('Предизвикателството е отхвърлено успешно');
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function viewChallengeResults(challengeId) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.challengeResults.replace(':challengeId', challengeId)}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при зареждане на резултатите');
        }

        const challenge = await response.json();
        displayChallengeResults(challenge);
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

function displayChallengeResults(challenge) {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <h3>Резултати от предизвикателството</h3>
            <p>Създател: ${challenge.creatorName}</p>
            <p>Опонент: ${challenge.opponentName}</p>
            <p>Период: ${new Date(challenge.startDate).toLocaleDateString('bg-BG')} - 
                      ${new Date(challenge.endDate).toLocaleDateString('bg-BG')}</p>
            <div class="results-container">
                ${challenge.results && challenge.results.length > 0 ? 
                    challenge.results.map(result => `
                        <div class="result-card">
                            <strong>${result.username}</strong>
                            <p>Начално тегло: ${result.initialWeight.toFixed(1)} кг</p>
                            ${result.finalWeight ? `
                                <p>Крайно тегло: ${result.finalWeight.toFixed(1)} кг</p>
                                <p>Прогрес: ${result.progress.toFixed(2)}%</p>
                            ` : '<p>Все още няма краен резултат</p>'}
                        </div>
                    `).join('') : 
                    '<p>Все още няма резултати</p>'
                }
            </div>
            <button onclick="this.closest('.modal').remove()">Затвори</button>
        </div>
    `;
    document.body.appendChild(modal);
}

// Помощни функции
function translateStatus(status) {
    const translations = {
        'pending': 'Изчакваща',
        'accepted': 'Приета',
        'rejected': 'Отхвърлена'
    };
    return translations[status] || status;
}

// Навигационни функции
async function showUsers() {
    try {
        console.log('Зареждане на компонента social...'); // Debug log
        await loadComponent('social');
        
        const usersContainer = document.getElementById('usersContainer');
        const friendsContainer = document.getElementById('friendsContainer');
        const challengesContainer = document.getElementById('challengesContainer');
        
        if (!usersContainer) {
            throw new Error('Не е намерен контейнер за потребители');
        }
        
        console.log('Показване на контейнера за потребители...'); // Debug log
        usersContainer.style.display = 'block';
        if (friendsContainer) friendsContainer.style.display = 'none';
        if (challengesContainer) challengesContainer.style.display = 'none';
        
        console.log('Зареждане на потребители...'); // Debug log
        await loadUsers();
    } catch (error) {
        console.error('Грешка при показване на потребители:', error);
        alert('Възникна грешка при зареждане на потребителите: ' + error.message);
    }
}

async function showFriends() {
    try {
        console.log('Зареждане на компонента social...'); // Debug log
        await loadComponent('social');
        
        const friendsContainer = document.getElementById('friendsContainer');
        const usersContainer = document.getElementById('usersContainer');
        const challengesContainer = document.getElementById('challengesContainer');
        
        if (!friendsContainer) {
            throw new Error('Не е намерен контейнер за приятели');
        }
        
        console.log('Показване на контейнера за приятели...'); // Debug log
        if (usersContainer) usersContainer.style.display = 'none';
        if (challengesContainer) challengesContainer.style.display = 'none';
        friendsContainer.style.display = 'block';
        
        console.log('Зареждане на приятели...'); // Debug log
        await loadFriends();
    } catch (error) {
        console.error('Грешка при показване на приятели:', error);
        alert('Възникна грешка при зареждане на приятелите: ' + error.message);
    }
}

async function showChallenges() {
    try {
        console.log('Зареждане на компонента social...'); // Debug log
        await loadComponent('social');
        
        const challengesContainer = document.getElementById('challengesContainer');
        const usersContainer = document.getElementById('usersContainer');
        const friendsContainer = document.getElementById('friendsContainer');
        
        if (!challengesContainer) {
            throw new Error('Не е намерен контейнер за предизвикателства');
        }
        
        console.log('Показване на контейнера за предизвикателства...'); // Debug log
        if (usersContainer) usersContainer.style.display = 'none';
        if (friendsContainer) friendsContainer.style.display = 'none';
        challengesContainer.style.display = 'block';
        
        console.log('Зареждане на предизвикателства...'); // Debug log
        await loadChallenges();
    } catch (error) {
        console.error('Грешка при показване на предизвикателства:', error);
        alert('Възникна грешка при зареждане на предизвикателствата: ' + error.message);
    }
}

function showNewChallengeForm(friendId) {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <h3>Ново предизвикателство</h3>
            <form id="newChallengeForm" onsubmit="createChallenge(event)">
                <input type="hidden" id="opponentId" value="${friendId}">
                <div class="form-group">
                    <label for="startDate">Начална дата:</label>
                    <input type="date" id="startDate" required>
                </div>
                <div class="form-group">
                    <label for="endDate">Крайна дата:</label>
                    <input type="date" id="endDate" required>
                </div>
                <div class="button-group">
                    <button type="submit">Създай</button>
                    <button type="button" onclick="this.closest('.modal').remove()">Отказ</button>
                </div>
            </form>
        </div>
    `;
    document.body.appendChild(modal);
}

async function createChallenge(event) {
    event.preventDefault();
    
    const opponentId = document.getElementById('opponentId').value;
    const startDate = document.getElementById('startDate').value;
    const endDate = document.getElementById('endDate').value;

    if (!opponentId || !startDate || !endDate) {
        alert('Моля, попълнете всички полета');
        return;
    }

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.challenges}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': localStorage.getItem('token')
            },
            body: JSON.stringify({
                opponentId: parseInt(opponentId),
                startDate: new Date(startDate).toISOString(),
                endDate: new Date(endDate).toISOString()
            })
        });

        if (!response.ok) {
            throw new Error('Грешка при създаване на предизвикателството');
        }

        document.querySelector('.modal').remove();
        await loadChallenges();
        alert('Предизвикателството е създадено успешно');
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

// Обновяваме функцията за показване на приятели, за да включва бутони за действие
function displayFriendRequests(requests) {
    const container = document.getElementById('friendRequests');
    container.innerHTML = '';

    requests.forEach(request => {
        console.log('Friend request object:', request); // Debug log
        
        const requestElement = document.createElement('div');
        requestElement.className = 'friend-request';
        requestElement.innerHTML = `
            <div class="friend-info">
                <h3>${request.username}</h3>
                <p>Статус: ${request.status}</p>
                <p>Прогрес: ${request.progress || 0}%</p>
            </div>
            <div class="friend-actions">
                <button onclick="acceptFriendRequest('${request.friendshipId}')" class="accept-btn">Приеми</button>
                <button onclick="rejectFriendRequest('${request.friendshipId}')" class="reject-btn">Отхвърли</button>
            </div>
        `;
        container.appendChild(requestElement);
    });
} 
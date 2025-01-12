// Глобална променлива за съхранение на предизвикателствата
let allChallenges = [];

// Функции за управление на предизвикателства
async function loadChallenges() {
    if (!isAuthenticated()) return;

    try {
        const response = await api.get(config.endpoints.challenges);

        if (response.error) {
            throw new Error(response.error);
        }

        allChallenges = response;

        updateChallengesList(allChallenges);
    } catch (error) {
        console.error('Error:', error);
        if (isAuthenticated()) {
            alert(error.message);
        }
    }
}

function updateChallengesList(challenges) {
    const container = document.getElementById('challengesList');
    if (!container) return;

    container.innerHTML = `
        <div class="challenges-header">
            <h3>Активни предизвикателства</h3>
            <button onclick="showCreateChallengeForm()">Ново предизвикателство</button>
        </div>
        <div id="createChallengeForm" style="display: none;">
            <form onsubmit="createChallenge(event)">
                <div class="form-group">
                    <label for="opponent">Противник:</label>
                    <select id="opponent" required>
                        <option value="">Изберете противник</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="startDate">Начална дата:</label>
                    <input type="date" id="startDate" required>
                </div>
                <div class="form-group">
                    <label for="endDate">Крайна дата:</label>
                    <input type="date" id="endDate" required>
                </div>
                <button type="submit">Създай</button>
                <button type="button" onclick="hideCreateChallengeForm()">Отказ</button>
            </form>
        </div>
        <div class="challenges-list">
    `;

    if (!challenges || challenges.length === 0) {
        container.innerHTML += '<p>Няма активни предизвикателства</p>';
        return;
    }

    challenges.forEach(challenge => {
        const startDate = new Date(challenge.startDate).toLocaleDateString('bg-BG');
        const endDate = new Date(challenge.endDate).toLocaleDateString('bg-BG');
        
        container.innerHTML += `
            <div class="challenge-card" data-id="${challenge.id}">
                <div class="challenge-header">
                    <h4>${challenge.creatorName} vs ${challenge.opponentName}</h4>
                    <span class="challenge-status ${challenge.status}">${getStatusText(challenge.status)}</span>
                </div>
                <div class="challenge-dates">
                    <p>От: ${startDate}</p>
                    <p>До: ${endDate}</p>
                </div>
                ${getChallengeActions(challenge)}
                ${challenge.status === 'completed' ? getChallengeResults(challenge) : ''}
            </div>
        `;
    });

    container.innerHTML += '</div>';
}

function getStatusText(status) {
    switch (status) {
        case 'pending': return 'Очаква отговор';
        case 'active': return 'Активно';
        case 'completed': return 'Завършено';
        case 'rejected': return 'Отхвърлено';
        default: return status;
    }
}

function getChallengeActions(challenge) {
    const currentUserId = getCurrentUserId();
    
    if (challenge.status === 'pending' && challenge.opponentId === currentUserId) {
        return `
            <div class="challenge-actions">
                <button onclick="acceptChallenge(${challenge.id})">Приеми</button>
                <button onclick="rejectChallenge(${challenge.id})">Отхвърли</button>
            </div>
        `;
    }
    
    if (challenge.status === 'active') {
        return `
            <div class="challenge-actions">
                <button onclick="viewChallengeResults(${challenge.id})">Виж прогрес</button>
            </div>
        `;
    }
    
    return '';
}

function getChallengeResults(challenge) {
    if (!challenge.results) return '';
    
    let resultsHtml = '<div class="challenge-results">';
    challenge.results.forEach(result => {
        const progress = result.progress ? result.progress.toFixed(2) : '0.00';
        resultsHtml += `
            <div class="result-card">
                <h5>${result.username}</h5>
                <p>Начално тегло: ${result.initialWeight} кг</p>
                <p>Крайно тегло: ${result.finalWeight || '-'} кг</p>
                <p>Прогрес: ${progress}%</p>
            </div>
        `;
    });
    resultsHtml += '</div>';
    
    return resultsHtml;
}

async function showCreateChallengeForm() {
    const form = document.getElementById('createChallengeForm');
    if (!form) return;
    
    // Зареждаме списъка с приятели
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.friends}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при зареждане на приятелите');
        }

        const friends = await response.json();
        const select = document.getElementById('opponent');
        select.innerHTML = '<option value="">Изберете противник</option>';
        
        friends
            .filter(friend => friend.status === 'accepted')
            .forEach(friend => {
                select.innerHTML += `<option value="${friend.id}">${friend.username}</option>`;
            });

        form.style.display = 'block';
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

function hideCreateChallengeForm() {
    const form = document.getElementById('createChallengeForm');
    if (form) {
        form.style.display = 'none';
    }
}

async function createChallenge(e) {
    e.preventDefault();
    
    const opponentId = document.getElementById('opponent').value;
    const startDate = document.getElementById('startDate').value;
    const endDate = document.getElementById('endDate').value;

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.challenges}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json; charset=utf-8',
                'Authorization': localStorage.getItem('token')
            },
            body: JSON.stringify({
                opponentId: parseInt(opponentId),
                startDate,
                endDate
            })
        });

        if (!response.ok) {
            throw new Error('Грешка при създаване на предизвикателството');
        }

        hideCreateChallengeForm();
        await loadChallenges();
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
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
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function rejectChallenge(challengeId) {
    if (!confirm('Сигурни ли сте, че искате да отхвърлите предизвикателството?')) {
        return;
    }

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
        showChallengeResults(challenge);
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

function showChallengeResults(challenge) {
    const container = document.getElementById('challengesList');
    if (!container) return;

    const startDate = new Date(challenge.startDate).toLocaleDateString('bg-BG');
    const endDate = new Date(challenge.endDate).toLocaleDateString('bg-BG');

    container.innerHTML = `
        <div class="challenge-details">
            <h3>Резултати от предизвикателството</h3>
            <button onclick="loadChallenges()">Назад</button>
            
            <div class="challenge-card">
                <div class="challenge-header">
                    <h4>${challenge.creatorName} vs ${challenge.opponentName}</h4>
                    <span class="challenge-status ${challenge.status}">${getStatusText(challenge.status)}</span>
                </div>
                <div class="challenge-dates">
                    <p>От: ${startDate}</p>
                    <p>До: ${endDate}</p>
                </div>
                ${getChallengeResults(challenge)}
            </div>
        </div>
    `;
}

// Помощни функции
function getCurrentUserId() {
    
    const user = authState.user;
    if (!user) return null;
    return user.id;
}

// Функция за показване на секцията с предизвикателства
async function showChallenges() {
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }

    // showSection('challengesSection');
    loadComponent('challenges');
    if (allChallenges.length === 0) {
        await loadChallenges();
    } else {
        updateChallengesList(allChallenges);
    }
}

// Функция за показване на формата за ново предизвикателство
window.showNewChallengeForm = function(friendId) {
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

// Функция за създаване на ново предизвикателство
async function createChallenge(event) {
    event.preventDefault();
    
    const opponentId = document.getElementById('opponent').value;
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
                'Content-Type': 'application/json; charset=utf-8',
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
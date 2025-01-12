// Функции за управление на теглото
async function addWeight() {
    if (!isAuthenticated()) return;

    const weight = parseFloat(document.getElementById('weight').value);
    const weightDate = document.getElementById('weightDate').value;
    
    if (!weight || !weightDate) {
        alert('Моля, попълнете всички полета');
        return;
    }

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.weight}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json; charset=utf-8',
                'Authorization': localStorage.getItem('token')
            },
            body: JSON.stringify({
                weight,
                createdAt: new Date(weightDate).toISOString()
            })
        });

        if (response.ok) {
            document.getElementById('weight').value = '';
            document.getElementById('weightDate').value = '';
            await showStats();
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при запазване на теглото');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Възникна грешка при комуникацията със сървъра');
    }
}

async function deleteWeight(weightId) {
    if (!isAuthenticated()) return;

    // Проверяваме дали имаме валидно ID
    if (!weightId) {
        console.error('No weight ID provided');
        alert('Грешка: Невалиден запис');
        return;
    }

    if (!confirm('Сигурни ли сте, че искате да изтриете този запис?')) {
        return;
    }

    try {
        console.log('Deleting weight with ID:', weightId);
        const response = await fetch(`${config.apiUrl}${config.endpoints.weightDelete.replace(':id', weightId)}`, {
            method: 'DELETE',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Грешка при изтриване на записа');
        }

        // Презареждаме статистиката след успешно изтриване
        await showStats();
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

async function loadStats() {
    if (!isAuthenticated()) return;

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.weightStats}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (response.ok) {
            const data = await response.json();
            updateStatsDisplay(data);
            updateHistoryDisplay(data.history);
        } else {
            const data = await response.json();
            // Не показваме alert при грешка, ако потребителят не е логнат
            if (isAuthenticated()) {
                alert(data.error || 'Грешка при зареждане на статистиката');
            }
        }
    } catch (error) {
        console.error('Error:', error);
        // Не показваме alert при грешка, ако потребителят не е логнат
        if (isAuthenticated()) {
            alert('Възникна грешка при комуникацията със сървъра');
        }
    }
}

// Помощни функции
function updateStatsDisplay(data) {
    document.getElementById('currentWeight').textContent = data.currentWeight ? data.currentWeight.toFixed(1) : '-';
    document.getElementById('initialWeight').textContent = data.initialWeight ? data.initialWeight.toFixed(1) : '-';
    document.getElementById('totalProgress').textContent = data.totalProgress ? data.totalProgress.toFixed(2) : '-';
    document.getElementById('dailyProgress').textContent = data.dailyProgress ? data.dailyProgress.toFixed(2) : '-';
    document.getElementById('bmi').textContent = data.bmi ? data.bmi.toFixed(1) : '-';
}

function updateHistoryDisplay(history) {
    const historyContainer = document.getElementById('weightHistory');
    if (!historyContainer) return;

    historyContainer.innerHTML = '';
    
    if (!history || history.length === 0) {
        historyContainer.innerHTML = '<p>Все още няма записи</p>';
        return;
    }

    history.forEach(record => {
        // Проверяваме дали record съществува и има валидни данни
        if (!record) {
            console.error('Missing record');
            return;
        }

        // Ако ID-то е 0 или undefined, използваме fallback стойност
        const recordId = record.id || record.ID; // Проверяваме и двата варианта
        if (!recordId) {
            console.error('Record without ID:', record);
            return;
        }

        const date = new Date(record.createdAt).toLocaleDateString('bg-BG');
        const weight = record.weight ? parseFloat(record.weight).toFixed(1) : '0.0';
        
        historyContainer.innerHTML += `
            <div class="history-item" data-id="${recordId}">
                <div>
                    <strong>${weight} кг</strong>
                    <span>${date}</span>
                </div>
                <button class="delete-button" onclick="deleteWeight(${recordId})">
                    Изтрий
                </button>
            </div>
        `;
    });
}

// Навигационни функции
async function showWeightForm() {
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }

    await loadComponent('weight');
    document.getElementById('weightForm').style.display = 'block';
    document.getElementById('statsContainer').style.display = 'none';
    document.getElementById('historyContainer').style.display = 'none';
    const now = new Date();
    now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
    document.getElementById('weightDate').value = now.toISOString().slice(0, 16);
}

async function showStats() {
    if (!isAuthenticated()) {
        loadComponent('auth');
        return;
    }

    await loadComponent('weight');
    document.getElementById('weightForm').style.display = 'none';
    document.getElementById('statsContainer').style.display = 'block';
    document.getElementById('historyContainer').style.display = 'block';
    await loadStats();
} 
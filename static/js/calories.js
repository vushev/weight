// Зареждане на настройките за калории
async function loadCalorieSettings() {
    if (!isAuthenticated()) return;

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.calorieSettings}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (response.status === 404) {
            // Ако няма настройки, показваме формата празна
            showCalorieSettingsForm({});
            return;
        }

        if (!response.ok) {
            throw new Error('Грешка при зареждане на настройките');
        }

        const settings = await response.json();
        showCalorieSettingsForm(settings);
    } catch (error) {
        console.error('Error:', error);
        if (isAuthenticated()) {
            alert(error.message);
        }
    }
}

// Запазване на настройките
async function saveCalorieSettings(e) {
    e.preventDefault();
    
    const form = e.target;
    const settings = {
        gender: form.gender.value,
        age: parseInt(form.age.value),
        activityLevel: form.activityLevel.value,
        goal: form.goal.value
    };

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.calorieSettings}`, {
            method: 'PUT',
            headers: {
                'Authorization': localStorage.getItem('token'),
                'Content-Type': 'application/json; charset=utf-8'
            },
            body: JSON.stringify(settings)
        });

        if (!response.ok) {
            throw new Error('Грешка при запазване на настройките');
        }

        alert('Настройките са запазени успешно');
        await loadCalorieStats();
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

// Зареждане на статистика за калориите
async function loadCalorieStats() {
    if (!isAuthenticated()) return;

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.calorieStats}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при зареждане на статистиката');
        }

        const stats = await response.json();
        showCalorieStats(stats);
    } catch (error) {
        console.error('Error:', error);
        if (isAuthenticated()) {
            alert(error.message);
        }
    }
}

// Зареждане на приема на калории
async function loadCalorieIntake(date = null) {
    if (!isAuthenticated()) return;

    try {
        // const url = new URL(`${config.apiUrl}${config.endpoints.calorieIntake}`);
        // if (date) {
        //     url.searchParams.append('date', date);
        // }
        const url = config.apiUrl + config.endpoints.calorieIntake + (date ? '?date=' + date : '');

        const response = await fetch(url, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при зареждане на приема на калории');
        }

        const intakes = await response.json();
        showCalorieIntake(intakes);
    } catch (error) {
        console.error('Error:', error);
        if (isAuthenticated()) {
            alert(error.message);
        }
    }
}

// Добавяне на прием на храна
async function addFoodEntry(event) {
    event.preventDefault();
    
    const form = event.target;
    const entry = {
        name: form.name.value,
        mealTypeId: parseInt(form.mealType.value),
        calories: parseFloat(form.calories.value),
        // protein: parseFloat(form.protein.value || 0),
        // carbs: parseFloat(form.carbs.value || 0),
        // fat: parseFloat(form.fat.value || 0),
        notes: form.notes.value || ''
    };

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.addFoodEntry}`, {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token'),
                'Content-Type': 'application/json; charset=utf-8'
            },
            body: JSON.stringify(entry)
        });

        if (!response.ok) {
            throw new Error('Грешка при добавяне на храна');
        }

        form.reset();
        await Promise.all([
            loadCalorieStats(),
            loadCalorieIntake()
        ]);

        await loadCalorieIntake();
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

// Показване на формата за добавяне на храна
function showFoodEntryForm() {
    const container = document.getElementById('addFoodForm');
    if (!container) return;

    container.innerHTML = `
        <h3>Добави храна</h3>
        <form onsubmit="addFoodEntry(event)">
            <div class="form-group">
                <label>Име на храната:</label>
                <input type="text" name="name" required>
            </div>
            <div class="form-group">
                <label>Тип хранене:</label>
                <select name="mealType" required>
                    <option value="">Изберете тип хранене</option>
                    <option value="1">Закуска</option>
                    <option value="2">Обяд</option>
                    <option value="3">Вечеря</option>
                    <option value="4">Междинно хранене</option>
                </select>
            </div>
            <div class="form-group">
                <label>Калории:</label>
                <input type="number" name="calories" step="0.1" required>
            </div>
            <div class="form-group">
                <label>Протеини (г):</label>
                <input type="number" name="protein" step="0.1">
            </div>
            <div class="form-group">
                <label>Въглехидрати (г):</label>
                <input type="number" name="carbs" step="0.1">
            </div>
            <div class="form-group">
                <label>Мазнини (г):</label>
                <input type="number" name="fat" step="0.1">
            </div>
            <div class="form-group">
                <label>Бележки:</label>
                <textarea name="notes"></textarea>
            </div>
            <button type="submit">Добави</button>
        </form>
    `;
}

// Показване на записите за храна
function showFoodEntries(entries) {
    const container = document.getElementById('foodEntries');
    if (!container) return;

    if (!entries || entries.length === 0) {
        container.innerHTML = '<p>Няма записи за храна</p>';
        return;
    }

    let html = '<div class="entries-list">';
    entries.forEach(entry => {
        html += `
            <div class="entry-card">
                <div class="entry-header">
                    <h4>${entry.name}</h4>
                    <span class="meal-type">${entry.mealType}</span>
                </div>
                <div class="entry-details">
                    <p><strong>Калории:</strong> ${entry.calories} kcal</p>
                    <p><strong>Протеини:</strong> ${entry.protein}г</p>
                    <p><strong>Въглехидрати:</strong> ${entry.carbs}г</p>
                    <p><strong>Мазнини:</strong> ${entry.fat}г</p>
                    ${entry.notes ? `<p><strong>Бележки:</strong> ${entry.notes}</p>` : ''}
                    <p class="entry-time">${new Date(entry.time).toLocaleString()}</p>
                </div>
                <button onclick="deleteFoodEntry(${entry.id})">Изтрий</button>
            </div>
        `;
    });
    html += '</div>';
    container.innerHTML = html;
}

// Изтриване на запис за калории
async function deleteCalorieIntake(id) {
    if (!confirm('Сигурни ли сте, че искате да изтриете този запис?')) {
        return;
    }

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.calorieIntakeDelete.replace(':id', id)}`, {
            method: 'DELETE',
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (!response.ok) {
            throw new Error('Грешка при изтриване на записа');
        }

        await Promise.all([
            loadCalorieStats(),
            loadCalorieIntake()
        ]);
    } catch (error) {
        console.error('Error:', error);
        alert(error.message);
    }
}

// Показване на формата за настройки
function showCalorieSettingsForm(settings) {
    const container = document.getElementById('calorieSettings');
    if (!container) return;

    container.innerHTML = `
        <h3>Настройки за калории</h3>
        <form id="calorieSettingsForm" onsubmit="saveCalorieSettings(event)">
            <div class="form-group">
                <label>Пол:</label>
                <select name="gender" required>
                    <option value="">Изберете пол</option>
                    <option value="male" ${settings.gender === 'male' ? 'selected' : ''}>Мъж</option>
                    <option value="female" ${settings.gender === 'female' ? 'selected' : ''}>Жена</option>
                </select>
            </div>
            <div class="form-group">
                <label>Възраст:</label>
                <input type="number" name="age" value="${settings.age || ''}" required>
            </div>
            <div class="form-group">
                <label>Ниво на активност:</label>
                <select name="activityLevel" required>
                    <option value="">Изберете ниво на активност</option>
                    <option value="sedentary" ${settings.activityLevel === 'sedentary' ? 'selected' : ''}>Заседнал начин на живот</option>
                    <option value="light" ${settings.activityLevel === 'light' ? 'selected' : ''}>Лека активност</option>
                    <option value="moderate" ${settings.activityLevel === 'moderate' ? 'selected' : ''}>Умерена активност</option>
                    <option value="active" ${settings.activityLevel === 'active' ? 'selected' : ''}>Активен начин на живот</option>
                    <option value="very_active" ${settings.activityLevel === 'very_active' ? 'selected' : ''}>Много активен начин на живот</option>
                </select>
            </div>
            <div class="form-group">
                <label>Цел:</label>
                <select name="goal" required>
                    <option value="">Изберете цел</option>
                    <option value="maintain" ${settings.goal === 'maintain' ? 'selected' : ''}>Поддържане на тегло</option>
                    <option value="lose" ${settings.goal === 'lose' ? 'selected' : ''}>Отслабване</option>
                    <option value="gain" ${settings.goal === 'gain' ? 'selected' : ''}>Качване</option>
                </select>
            </div>
            <button type="submit">Запази настройките</button>
        </form>
    `;
}

// Показване на статистиката
function showCalorieStats(stats) {
    const container = document.getElementById('calorieStats');
    if (!container) return;

    container.innerHTML = `
        <h3>Статистика за калориите</h3>
        <div class="stats-grid">
            <div class="stat-card">
                <h4>Базален метаболизъм (BMR)</h4>
                <p>${stats.bmr || 0} кал</p>
            </div>
            <div class="stat-card">
                <h4>Калории за поддържане</h4>
                <p>${stats.dailyNeeds.maintenance || 0} кал</p>
            </div>
            <div class="stat-card">
                <h4>Целеви калории</h4>
                <p>${stats.targetCalories || 0} кал</p>
            </div>
        </div>
        <div class="daily-stats">
            <h4>Днешен ден</h4>
            <p>Приети калории: ${stats.intakeAnalysis.totalCurrentDayIntake || 0}</p>
            <p>Изгорени калории: ${stats.intakeAnalysis.dailyBurned || 0}</p>
            <p>Нетни калории: ${stats.intakeAnalysis.netCalories || 0}</p>
        </div>
    `;
}

// Показване на приема на калории
async function showCalorieIntake(intakes) {
    const container = document.getElementById('calorieIntake');
    if (!container) return;
console.warn(intakes);
    let foodHtml = '<h4>Храна</h4>';
    let exerciseHtml = '<h4>Физическа активност</h4>';

    intakes.foodEntries.forEach(intake => {
        const html = `
            <div class="intake-card">
                <div>
                    <strong>${intake.calories} кал</strong>
                    ${intake.meal ? `<p>Хранене: ${translateMeal(intake.meal)}</p>` : ''}
                    ${intake.notes ? `<p>Бележки: ${intake.notes}</p>` : ''}
                    <small>${new Date(intake.time).toLocaleString()}</small>
                </div>
                <button onclick="deleteCalorieIntake(${intake.id})">Изтрий</button>
            </div>
        `;

        if (intake.type === 'food') {
            foodHtml += html;
        } else {
            exerciseHtml += html;
        }
    });

    let request = await fetch(`${config.apiUrl}${config.endpoints.mealTypes}`, {
        headers: {
            'Authorization': localStorage.getItem('token')
        }
    });
    let mealTypes = await request.json();
    console.warn(mealTypes);

    container.innerHTML = `
        <h3>Дневен прием</h3>
        <div class="intake-form">
            <form onsubmit="addFoodEntry(event)">
                <div class="form-group" id="mealField">
                    <label>Хранене:</label>
                    <select name="mealType">
                        ${mealTypes.map(mealType => `<option value="${mealType.id}">${mealType.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label>Калории:</label>
                    <input type="number" name="calories" required>
                </div>
                <div class="form-group">
                    <label>Бележки:</label>
                    <input type="text" name="notes">
                </div>
                <button type="submit">Добави</button>
            </form>
        </div>
        <div class="intake-list">
            ${foodHtml}
            ${exerciseHtml}
        </div>
    `;
}

// Помощна функция за превод на типа хранене
function translateMeal(meal) {
    const translations = {
        breakfast: 'Закуска',
        lunch: 'Обяд',
        dinner: 'Вечеря',
        snack: 'Междинно хранене'
    };
    return translations[meal] || meal;
}

// Помощна функция за показване/скриване на полето за хранене
function toggleMealField(select) {
    const mealField = document.getElementById('mealField');
    if (mealField) {
        mealField.style.display = select.value === 'food' ? 'block' : 'none';
    }
}

// Инициализация
document.addEventListener('UserLoggedIn', () => {
    loadCalorieSettings();
    loadCalorieStats();
    loadCalorieIntake();
});
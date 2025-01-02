// Функции за настройки
async function loadUserSettings() {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.userSettings}`, {
            headers: {
                'Authorization': localStorage.getItem('token')
            }
        });

        if (response.ok) {
            const data = await response.json();
            updateSettingsForm(data);
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Грешка при зареждане на настройките');
    }
}

async function saveSettings() {
    const settings = {
        firstName: document.getElementById('firstName').value,
        lastName: document.getElementById('lastName').value,
        email: document.getElementById('email').value,
        age: parseInt(document.getElementById('age').value),
        height: parseInt(document.getElementById('height').value),
        gender: document.getElementById('gender').value,
        targetWeight: parseFloat(document.getElementById('targetWeight').value),
        isVisible: document.getElementById('isVisible').checked
    };

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.userSettings}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': localStorage.getItem('token')
            },
            body: JSON.stringify(settings)
        });

        if (response.ok) {
            await updateVisibility(settings.isVisible);
            alert('Настройките са запазени успешно');
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при запазване на настройките');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Грешка при комуникацията със сървъра');
    }
}

async function changePassword() {
    const currentPassword = document.getElementById('currentPassword').value;
    const newPassword = document.getElementById('newPassword').value;

    if (!currentPassword || !newPassword) {
        alert('Моля, попълнете всички полета');
        return;
    }

    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.changePassword}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': localStorage.getItem('token')
            },
            body: JSON.stringify({
                currentPassword,
                newPassword
            })
        });

        if (response.ok) {
            alert('Паролата е променена успешно');
            document.getElementById('currentPassword').value = '';
            document.getElementById('newPassword').value = '';
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при промяна на паролата');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Грешка при комуникацията със сървъра');
    }
}

// Помощни функции
function updateSettingsForm(data) {
    document.getElementById('firstName').value = data.firstName || '';
    document.getElementById('lastName').value = data.lastName || '';
    document.getElementById('email').value = data.email || '';
    document.getElementById('age').value = data.age || '';
    document.getElementById('height').value = data.height || '';
    document.getElementById('gender').value = data.gender || '';
    document.getElementById('targetWeight').value = data.targetWeight || '';
    document.getElementById('isVisible').checked = data.isVisible;
}

// Навигационни функции
async function showSettings() {
    await loadComponent('settings');
    document.getElementById('settingsForm').style.display = 'block';
    document.getElementById('changePasswordForm').style.display = 'none';
    await loadUserSettings();
}

function showChangePassword() {
    document.getElementById('settingsForm').style.display = 'none';
    document.getElementById('changePasswordForm').style.display = 'block';
}

function showSettingsForm() {
    document.getElementById('settingsForm').style.display = 'block';
    document.getElementById('changePasswordForm').style.display = 'none';
}

// Добавяме нова функция за обновяване на видимостта
async function updateVisibility(isVisible) {
    try {
        const response = await fetch(`${config.apiUrl}${config.endpoints.updateVisibility}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': localStorage.getItem('token')
            },
            body: JSON.stringify({ isVisible })
        });

        if (response.ok) {
            alert('Видимостта на профила е обновена успешно');
        } else {
            const data = await response.json();
            alert(data.error || 'Грешка при обновяване на видимостта');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Грешка при комуникацията със сървъра');
    }
} 
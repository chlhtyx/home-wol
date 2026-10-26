document.getElementById('loginForm').addEventListener('submit', async (event) => {
    event.preventDefault();
    const button = document.getElementById('loginButton');
    const status = document.getElementById('loginStatus');
    if (button.disabled) return;
    button.disabled = true;
    button.textContent = '登录中…';
    status.textContent = '';
    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            credentials: 'same-origin',
            headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest' },
            body: JSON.stringify({
                username: document.getElementById('username').value,
                password: document.getElementById('password').value
            })
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || '登录失败，请重试');
        }
        document.getElementById('password').value = '';
        window.location.replace('/admin/');
    } catch (error) {
        status.textContent = error instanceof TypeError ? '连接失败，请检查网络后重试' : error.message;
        document.getElementById('password').value = '';
        document.getElementById('password').focus();
    } finally {
        button.disabled = false;
        button.textContent = '登录';
    }
});

// auth.js - 处理用户认证相关功能

// 存储用户信息和令牌
const Auth = {
    token: null,
    username: null,
    userId: null,

    // 从localStorage加载认证信息
    loadFromStorage: function() {
        this.token = localStorage.getItem('token');
        this.username = localStorage.getItem('username');
        this.userId = localStorage.getItem('userId');
        return this.isAuthenticated();
    },

    // 保存认证信息到localStorage
    saveToStorage: function() {
        if (this.token) {
            localStorage.setItem('token', this.token);
            localStorage.setItem('username', this.username);
            localStorage.setItem('userId', this.userId);
        }
    },

    // 清除认证信息
    clear: function() {
        this.token = null;
        this.username = null;
        this.userId = null;
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        localStorage.removeItem('userId');
    },

    // 检查是否已认证
    isAuthenticated: function() {
        return !!this.token;
    },

    // 登录
    login: async function(username, password) {
        try {
            // 简化版：在实际应用中，这里应该调用后端API进行验证
            // 这里我们使用模拟数据，实际项目中应该发送到服务器验证
            
            // 模拟API调用
            // const response = await fetch('/api/auth/login', {
            //     method: 'POST',
            //     headers: { 'Content-Type': 'application/json' },
            //     body: JSON.stringify({ username, password })
            // });
            // const data = await response.json();
            
            // 模拟成功响应
            const data = {
                success: true,
                token: 'mock-jwt-token-' + Math.random().toString(36).substring(2),
                userId: 'user-' + Math.random().toString(36).substring(2),
                username: username
            };

            if (data.success) {
                this.token = data.token;
                this.username = data.username;
                this.userId = data.userId;
                this.saveToStorage();
                return { success: true };
            } else {
                return { success: false, message: data.message || '登录失败' };
            }
        } catch (error) {
            console.error('登录错误:', error);
            return { success: false, message: '登录过程中发生错误' };
        }
    },

    // 注册
    register: async function(username, password) {
        try {
            // 简化版：在实际应用中，这里应该调用后端API进行注册
            // 这里我们使用模拟数据，实际项目中应该发送到服务器
            
            // 模拟API调用
            // const response = await fetch('/api/auth/register', {
            //     method: 'POST',
            //     headers: { 'Content-Type': 'application/json' },
            //     body: JSON.stringify({ username, password })
            // });
            // const data = await response.json();
            
            // 模拟成功响应
            const data = {
                success: true,
                message: '注册成功'
            };

            return { 
                success: data.success, 
                message: data.message || '注册成功' 
            };
        } catch (error) {
            console.error('注册错误:', error);
            return { success: false, message: '注册过程中发生错误' };
        }
    },

    // 登出
    logout: function() {
        this.clear();
    }
};

// 页面加载时初始化认证状态
document.addEventListener('DOMContentLoaded', function() {
    // 加载认证信息
    const isAuthenticated = Auth.loadFromStorage();
    
    // 根据认证状态显示相应界面
    if (isAuthenticated) {
        document.getElementById('auth-screen').classList.add('hidden');
        document.getElementById('lobby-screen').classList.remove('hidden');
        document.getElementById('user-name').textContent = Auth.username;
    }

    // 登录按钮事件
    document.getElementById('login-btn').addEventListener('click', async function() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        
        if (!username || !password) {
            alert('请输入用户名和密码');
            return;
        }
        
        const result = await Auth.login(username, password);
        
        if (result.success) {
            document.getElementById('auth-screen').classList.add('hidden');
            document.getElementById('lobby-screen').classList.remove('hidden');
            document.getElementById('user-name').textContent = Auth.username;
        } else {
            alert(result.message);
        }
    });

    // 注册按钮事件
    document.getElementById('register-btn').addEventListener('click', async function() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        
        if (!username || !password) {
            alert('请输入用户名和密码');
            return;
        }
        
        const result = await Auth.register(username, password);
        
        if (result.success) {
            alert('注册成功，请登录');
        } else {
            alert(result.message);
        }
    });

    // 登出按钮事件
    document.getElementById('logout-btn').addEventListener('click', function() {
        Auth.logout();
        document.getElementById('lobby-screen').classList.add('hidden');
        document.getElementById('auth-screen').classList.remove('hidden');
    });
});
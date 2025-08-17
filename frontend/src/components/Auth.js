import React, { useState } from 'react';
import { authAPI } from '../api';

const Auth = ({ onLogin }) => {
  const [isLogin, setIsLogin] = useState(true);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (isLogin) {
        const response = await authAPI.login(username, password);
        const { token, user_id, username: userName } = response.data;
        
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify({ id: user_id, username: userName }));
        
        onLogin({ id: user_id, username: userName });
      } else {
        await authAPI.register(username, password);
        setError('');
        setIsLogin(true);
        setPassword('');
        alert('Регистрация успешна! Теперь войдите в систему.');
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Произошла ошибка');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-form">
      <h2>{isLogin ? 'Вход в систему' : 'Регистрация'}</h2>
      
      {error && <div className="error">{error}</div>}
      
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Имя пользователя:</label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            disabled={loading}
          />
        </div>
        
        <div className="form-group">
          <label>Пароль:</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={loading}
          />
        </div>
        
        <button type="submit" className="btn" disabled={loading}>
          {loading ? 'Загрузка...' : (isLogin ? 'Войти' : 'Зарегистрироваться')}
        </button>
      </form>
      
      <div className="auth-switch">
        <button 
          type="button" 
          onClick={() => {
            setIsLogin(!isLogin);
            setError('');
            setPassword('');
          }}
          disabled={loading}
        >
          {isLogin ? 'Нет аккаунта? Зарегистрироваться' : 'Есть аккаунт? Войти'}
        </button>
      </div>
    </div>
  );
};

export default Auth;

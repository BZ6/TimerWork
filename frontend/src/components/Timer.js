import React, { useState, useEffect } from 'react';
import { timerAPI } from '../api';

const Timer = ({ user, onLogout }) => {
  const [elapsedTime, setElapsedTime] = useState(0);
  const [status, setStatus] = useState('stopped');
  const [weekStart, setWeekStart] = useState(null);
  const [weekEnd, setWeekEnd] = useState(null);
  const [weekGoal, setWeekGoal] = useState(2400); // цель в минутах (40 часов по умолчанию)
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showGoalModal, setShowGoalModal] = useState(false);
  const [goalInput, setGoalInput] = useState('40'); // в часах для удобства
  const [goalUnit, setGoalUnit] = useState('hours');

  // Загрузка данных при монтировании компонента
  useEffect(() => {
    loadWorkWeek();
  }, []);

  // Убираем автоматическое обновление времени
  // useEffect(() => {
  //   const interval = setInterval(() => {
  //     if (status === 'running') {
  //       updateCurrentTime();
  //     }
  //   }, 1000);

  //   return () => clearInterval(interval);
  // }, [status]);

  // Сохранение времени в sessionStorage
  useEffect(() => {
    sessionStorage.setItem('elapsedTime', elapsedTime.toString());
    sessionStorage.setItem('status', status);
  }, [elapsedTime, status]);

  // Загрузка из sessionStorage при запуске
  useEffect(() => {
    const savedTime = sessionStorage.getItem('elapsedTime');
    const savedStatus = sessionStorage.getItem('status');
    
    if (savedTime) {
      setElapsedTime(parseInt(savedTime, 10));
    }
    if (savedStatus) {
      setStatus(savedStatus);
    }
  }, []);

  const loadWorkWeek = async () => {
    try {
      const response = await timerAPI.getWorkWeek();
      const workWeek = response.data.work_week;
      
      if (workWeek) {
        setElapsedTime(workWeek.elapsed_time);
        setStatus(workWeek.status);
        setWeekStart(new Date(workWeek.week_start));
        setWeekEnd(workWeek.week_end ? new Date(workWeek.week_end) : null);
        setWeekGoal(workWeek.week_goal_minutes || 2400);
      }
    } catch (err) {
      setError('Ошибка загрузки данных');
    }
  };

  const updateCurrentTime = async () => {
    try {
      const response = await timerAPI.getCurrentTime();
      setElapsedTime(response.data.elapsed_time);
      setStatus(response.data.status);
      showMessage('Время обновлено');
    } catch (err) {
      showMessage('Ошибка обновления времени', true);
    }
  };

  const showMessage = (message, isError = false) => {
    if (isError) {
      setError(message);
      setSuccess('');
    } else {
      setSuccess(message);
      setError('');
    }
    
    setTimeout(() => {
      setError('');
      setSuccess('');
    }, 3000);
  };

  const handleStartWeek = async () => {
    setShowGoalModal(true);
  };

  const handleStartWeekWithGoal = async () => {
    try {
      // Конвертируем цель в минуты
      let goalMinutes;
      if (goalUnit === 'hours') {
        goalMinutes = parseInt(goalInput) * 60;
      } else {
        goalMinutes = parseInt(goalInput);
      }

      if (goalMinutes <= 0) {
        showMessage('Цель должна быть больше 0', true);
        return;
      }

      await timerAPI.startWeek(goalMinutes);
      setStatus('running');
      setWeekStart(new Date());
      setWeekEnd(null);
      setElapsedTime(0);
      setWeekGoal(goalMinutes);
      setShowGoalModal(false);
      showMessage(`Рабочая неделя начата! Цель: ${formatTime(goalMinutes * 60)}`);
    } catch (err) {
      showMessage(err.response?.data?.error || 'Ошибка начала недели', true);
    }
  };

  const handleEndWeek = async () => {
    try {
      await timerAPI.endWeek();
      setStatus('stopped');
      setWeekEnd(new Date());
      showMessage('Рабочая неделя завершена!');
    } catch (err) {
      showMessage(err.response?.data?.error || 'Ошибка завершения недели', true);
    }
  };

  const handlePause = async () => {
    try {
      await timerAPI.pauseTimer();
      setStatus('paused');
      showMessage('Таймер поставлен на паузу');
    } catch (err) {
      showMessage(err.response?.data?.error || 'Ошибка паузы', true);
    }
  };

  const handleResume = async () => {
    try {
      await timerAPI.resumeTimer();
      setStatus('running');
      showMessage('Таймер возобновлен');
    } catch (err) {
      showMessage(err.response?.data?.error || 'Ошибка возобновления', true);
    }
  };

  const formatTime = (seconds) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const formatDate = (date) => {
    if (!date) return 'Не установлено';
    return date.toLocaleString('ru-RU');
  };

  const getStatusText = () => {
    switch (status) {
      case 'running':
        return 'Работает';
      case 'paused':
        return 'Пауза';
      case 'stopped':
        return 'Остановлен';
      default:
        return 'Неизвестно';
    }
  };

  return (
    <div className="container">
      {error && <div className="error">{error}</div>}
      {success && <div className="success">{success}</div>}

      <div className="timer-display">
        <h1>{formatTime(elapsedTime)}</h1>
        <div className={`status ${status}`}>
          Статус: {getStatusText()}
        </div>
        
        {/* Информация о цели и прогрессе */}
        {weekGoal > 0 && (status === 'running' || status === 'paused') && (
          <div className="goal-progress">
            <div className="goal-info">
              <span>Цель недели: {formatTime(weekGoal * 60)}</span>
              <span>Прогресс: {Math.round((elapsedTime / (weekGoal * 60)) * 100)}%</span>
            </div>
            <div className="progress-bar">
              <div 
                className="progress-fill"
                style={{ 
                  width: `${Math.min((elapsedTime / (weekGoal * 60)) * 100, 100)}%`,
                  backgroundColor: elapsedTime >= weekGoal * 60 ? '#28a745' : '#007bff'
                }}
              ></div>
            </div>
          </div>
        )}
        
        <div style={{ marginTop: '15px', fontSize: '0.9em', color: '#666' }}>
          <div>Начало недели: {formatDate(weekStart)}</div>
          <div>Конец недели: {formatDate(weekEnd)}</div>
        </div>
      </div>

      <div className="controls">
        <button
          onClick={handleStartWeek}
          className="btn btn-success"
          disabled={status === 'running' || status === 'paused'}
        >
          Начало недели
        </button>

        <button
          onClick={handleEndWeek}
          className="btn btn-danger"
          disabled={status === 'stopped'}
        >
          Конец недели
        </button>

        <button
          onClick={handlePause}
          className="btn btn-warning"
          disabled={status !== 'running'}
        >
          Пауза
        </button>

        <button
          onClick={handleResume}
          className="btn btn-success"
          disabled={status !== 'paused'}
        >
          Продолжить
        </button>

        <button
          onClick={updateCurrentTime}
          className="btn btn-secondary"
          disabled={status === 'stopped'}
        >
          Обновить время
        </button>
      </div>

      {/* Модальное окно для установки цели */}
      {showGoalModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <h3>Установить цель на неделю</h3>
            <div className="goal-input-group">
              <input
                type="number"
                value={goalInput}
                onChange={(e) => setGoalInput(e.target.value)}
                min="1"
                placeholder="Введите цель"
              />
              <select 
                value={goalUnit} 
                onChange={(e) => setGoalUnit(e.target.value)}
              >
                <option value="hours">часов</option>
                <option value="minutes">минут</option>
              </select>
            </div>
            <div className="modal-buttons">
              <button 
                onClick={handleStartWeekWithGoal}
                className="btn btn-success"
              >
                Начать неделю
              </button>
              <button 
                onClick={() => setShowGoalModal(false)}
                className="btn btn-secondary"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Timer;

import React, { useState, useContext } from 'react';
import api from '../utils/api';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();
    const { login } = useContext(AuthContext);

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const response = await api.post('login', { username, password });
            login(response.data.token); // Используем контекст для обновления
            navigate('/');
        } catch (err) {
            setError(err.response?.data?.error || 'Ошибка входа');
            console.error('Login error:', err);
        }
    };

    const Back = () => {
    navigate('/');
  };

    return (
        <form onSubmit={handleSubmit}>
            <div className="mb-3">
                <label>Login:</label>
                <input type="text" value={username} onChange={(e) => setUsername(e.target.value)} className="form-control" required />
            </div>
            <div className="mb-3">
                <label>Password:</label>
                <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} className="form-control" required />
            </div>
            {error && <div className="alert alert-danger">{error}</div>}
            <button type="submit" className="btn btn-primary">Enter</button>
            <button onClick={Back} className="btn btn-primary">Back</button>
        </form>
    );
};

export default Login;
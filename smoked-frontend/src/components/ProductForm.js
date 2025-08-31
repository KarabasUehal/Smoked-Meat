import React, { useState, useEffect } from 'react';
import api from '../utils/api';
import { useParams, useNavigate } from 'react-router-dom';

const ProductForm = () => {
    const [meat, setMeat] = useState('');
    const [avail, setAvail] = useState(true);
    const [price, setPrice] = useState('');
    const [spice, setSpice] = useState({ recipe1: '', recipe2: '' });
    const [error, setError] = useState('');
    const navigate = useNavigate();
    const { id } = useParams();

    useEffect(() => {
        if (id) {
            fetchProduct();
        }
    }, [id]);

    const fetchProduct = async () => {
        try {
            const response = await api.get(`/product/${id}`);
            const { meat, avail, price, spice } = response.data;
            setMeat(meat);
            setAvail(avail);
            setPrice(price);
            setSpice(spice || { recipe1: '', recipe2: '' });
        } catch (err) {
            setError('Ошибка при загрузке продукта');
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const data = { meat, avail, price: parseFloat(price), spice };
            if (id) {
                await api.put(`/product/${id}`, data);
            } else {
                await api.post('/product', data);
            }
            navigate('/');
        } catch (err) {
            setError(err.response?.data?.error || 'Ошибка при сохранении продукта');
        }
    };

    const Back = () => {
    navigate('/');
  };

    return (
        <form onSubmit={handleSubmit}>
            <h2>{id ? 'Редактировать продукт' : 'Добавить продукт'}</h2>
            <div className="mb-3">
                <label>Meat:</label>
                <input
                    type="text"
                    value={meat}
                    onChange={(e) => setMeat(e.target.value)}
                    className="form-control"
                    style={{ backgroundColor: 'transparent', color: '#FFFF00', textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)' }}
                    required
                />
            </div>
            <div className="mb-3">
                <label>Availability:</label>
                <input
                    type="checkbox"
                    checked={avail}
                    onChange={(e) => setAvail(e.target.checked)}
                />
            </div>
            <div className="mb-3">
                <label>Price:</label>
                <input
                    type="number"
                    value={price}
                    onChange={(e) => setPrice(e.target.value)}
                    className="form-control"
                    style={{ backgroundColor: 'transparent', color: '#FFFF00', textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)' }}
                    required
                />
            </div>
            <div className="mb-3">
                <label>Spice 1:</label>
                <input
                    type="text"
                    value={spice.recipe1}
                    onChange={(e) => setSpice({ ...spice, recipe1: e.target.value })}
                    className="form-control"
                    style={{ backgroundColor: 'transparent', color: '#FFFF00', textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)' }}
                    required
                />
            </div>
            <div className="mb-3">
                <label>Spice 2:</label>
                <input
                    type="text"
                    value={spice.recipe2}
                    onChange={(e) => setSpice({ ...spice, recipe2: e.target.value })}
                    className="form-control"
                    style={{ backgroundColor: 'transparent', color: '#FFFF00', textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)' }}
                    required
                />
            </div>
            {error && <div className="alert alert-danger">{error}</div>}
            <button type="submit" className="btn btn-primary">
                Save
            </button>
            <button onClick={Back} className="btn btn-primary">
                Back
            </button>
        </form>
    );
};

export default ProductForm;
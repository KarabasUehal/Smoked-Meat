import React, { useState, useEffect } from 'react';
import api from '../utils/api';
import axios from 'axios';

const CalculatePrice = () => {
    const [products, setProducts] = useState([]);
    const [selectedProduct, setSelectedProduct] = useState('');
    const [quantity, setQuantity] = useState('');
    const [result, setResult] = useState(null);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchProducts = async () => {
            try {
                const response = await api.get('/assortment');
                setProducts(response.data);
                if (response.data.length > 0) {
                    setSelectedProduct(response.data[0].id.toString()); 
                }
            } catch (err) {
                setError('Ошибка при загрузке товаров');
            }
        };
        fetchProducts();
    }, []);

    const handleCalculate = async (e) => {
        e.preventDefault();
        setError('');
        setResult(null);

        // Валидация
        const parsedId = parseInt(selectedProduct);
        const parsedQuantity = parseFloat(quantity);

        if (!selectedProduct || isNaN(parsedId)) {
            setError('Выберите товар');
            return;
        }
        if (!quantity || isNaN(parsedQuantity) || parsedQuantity <= 0) {
            setError('Введите корректное количество (больше 0)');
            return;
        }

        try {
            const response = await api.post('/calculate-price', {
                id: parsedId,
                quantity: parsedQuantity,
            });
            setResult(response.data);
        } catch (err) {
            setError('Ошибка при расчете цены: ' + (err.response?.data?.error || err.message));
        }
    };

    return (
        <div className="mt-4">
            <h2>Choose product and quantity to calculate the price</h2>
            {error && <div className="alert alert-danger">{error}</div>}
            <form onSubmit={handleCalculate}>
                <div className="mb-3">
                    <label>Product:</label>
                    <select
                        className="form-control"
                        value={selectedProduct}
                        onChange={(e) => setSelectedProduct(e.target.value)}
                    >
                        {products.length === 0 && (
                            <option value="">No available products</option>
                        )}
                        {products.map((product) => (
                            <option key={product.id} value={product.id}>
                                {product.meat} ({product.price} rub./kg)
                            </option>
                        ))}
                    </select>
                </div>
                <div className="mb-3">
                    <label>Quantity (кг):</label>
                    <input
                        type="number"
                        step="0.1"
                        min="0"
                        className="form-control"
                        value={quantity}
                        onChange={(e) => setQuantity(e.target.value)}
                        required
                    />
                </div>
                <button type="submit" className="btn btn-primary">Calculate</button>
            </form>
            {result && (
                <div className="mt-3">
                    <h4>Your choise:</h4>
                    <p>{result.meat}</p>
                    <p>{result.quantity} kg</p>
                    <p>Final price: {result.total_price} rub.</p>
                </div>
            )}
        </div>
    );
};

export default CalculatePrice;
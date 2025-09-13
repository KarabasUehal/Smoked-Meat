import React, { createContext, useState, useEffect } from 'react';
import { emitter } from './events';

export const CartContext = createContext();

export const CartProvider = ({ children }) => {
    const [cart, setCart] = useState([]); 

    useEffect(() => {
        const handleAuthChange = () => {
            setCart([]);
            console.log('Cart cleared due to authChange event');
        };
        emitter.on('authChange', handleAuthChange);
        return () => emitter.off('authChange', handleAuthChange); // Очищаем подписку
    }, []);

    const addToCart = (product, quantity, selectedSpice) => {
        setCart((prevCart) => {
            const itemKey = `${product.id}-${selectedSpice}`; // Уникальный ключ: id + специя
            const existing = prevCart.find((item) => item.itemKey === itemKey);
            if (existing) {
                return prevCart.map((item) =>
                    item.itemKey === itemKey ? { ...item, quantity: item.quantity + quantity } : item
                );
            }
            return [
                ...prevCart,
                {
                    ...product,
                    quantity,
                    selectedSpice: selectedSpice || product.spice.recipe1 || product.spice.recipe2,
                    itemKey, // Уникальный ключ для идентификации
                },
            ];
        });
    };

    const removeFromCart = (itemKey) => {
        setCart((prevCart) => prevCart.filter((item) => item.itemKey !== itemKey));
    };

    const updateQuantity = (itemKey, quantity) => {
        setCart((prevCart) =>
            prevCart.map((item) =>
                item.itemKey === itemKey ? { ...item, quantity: Math.max(0, quantity) } : item
            )
        );
    };

    const updateSpice = (itemKey, selectedSpice) => {
        setCart((prevCart) => {
            const item = prevCart.find((item) => item.itemKey === itemKey);
            if (!item) return prevCart;

            const newItemKey = `${item.id}-${selectedSpice}`;
            if (prevCart.some((i) => i.itemKey === newItemKey)) {
                // Если такой продукт с этой специей уже есть, объединяем
                return prevCart
                    .map((i) =>
                        i.itemKey === newItemKey
                            ? { ...i, quantity: i.quantity + item.quantity }
                            : i
                    )
                    .filter((i) => i.itemKey !== itemKey);
            }
            return prevCart.map((i) =>
                i.itemKey === itemKey ? { ...i, selectedSpice, itemKey: newItemKey } : i
            );
        });
    };

    const clearCart = () => {
        setCart([]);
    };

    const getTotalPrice = () => {
        return cart.reduce((total, item) => total + item.price * item.quantity, 0);
    };

    return (
        <CartContext.Provider
            value={{ cart, addToCart, removeFromCart, updateQuantity, updateSpice, clearCart, getTotalPrice }}
        >
            {children}
        </CartContext.Provider>
    );
};
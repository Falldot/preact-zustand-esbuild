import React from 'react';
import create from 'zustand'

const useStore = create((set) => ({
    bears: 0,
    increasePopulation: () => set(state => ({ bears: state.bears + 1 })),
    removeAllBears: () => set({ bears: 0 })
}))

export default  function TextInput() {
    const bears = useStore(state => state)
    
    return (
        <div>
            <h1>{bears.bears}</h1>
            <button onClick={bears.increasePopulation}>Добавить</button>
        </div>
    );
}

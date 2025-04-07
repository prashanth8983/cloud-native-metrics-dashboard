import React, { useState } from 'react';
import { Calendar, ChevronLeft, ChevronRight } from 'lucide-react';
//import { useTheme } from '../../context/ThemeContext';
import { TimeRange } from '../../types/common';

interface DateRangePickerProps {
  timeRange: TimeRange;
  onChange: (range: TimeRange) => void;
  className?: string;
}

const DateRangePicker: React.FC<DateRangePickerProps> = ({
  timeRange,
  onChange,
  className = ''
}) => {
  //const { theme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [selecting, setSelecting] = useState<'start' | 'end'>('start');
  const [tempRange, setTempRange] = useState<TimeRange>(timeRange);
  const [currentMonth, setCurrentMonth] = useState<Date>(new Date(tempRange.start));

  const formatDate = (date: Date): string => {
    return date.toLocaleDateString(undefined, { 
      year: 'numeric', 
      month: 'short', 
      day: 'numeric' 
    });
  };

  const toggleDatepicker = () => {
    setIsOpen(!isOpen);
    if (!isOpen) {
      setTempRange(timeRange);
      setCurrentMonth(new Date(timeRange.start));
      setSelecting('start');
    }
  };

  const handleDateClick = (date: Date) => {
    if (selecting === 'start') {
      const newRange = { ...tempRange, start: date };
      setTempRange(newRange);
      setSelecting('end');
      setCurrentMonth(date);
    } else {
      // Ensure end date is not before start date
      if (date >= tempRange.start) {
        const newRange = { ...tempRange, end: date };
        setTempRange(newRange);
        onChange(newRange);
        setIsOpen(false);
      }
    }
  };

  const previousMonth = () => {
    const prevMonth = new Date(currentMonth);
    prevMonth.setMonth(currentMonth.getMonth() - 1);
    setCurrentMonth(prevMonth);
  };

  const nextMonth = () => {
    const nextMonth = new Date(currentMonth);
    nextMonth.setMonth(currentMonth.getMonth() + 1);
    setCurrentMonth(nextMonth);
  };

  const getDaysInMonth = (year: number, month: number) => {
    return new Date(year, month + 1, 0).getDate();
  };

  const getFirstDayOfMonth = (year: number, month: number) => {
    return new Date(year, month, 1).getDay();
  };

  const renderCalendar = () => {
    const year = currentMonth.getFullYear();
    const month = currentMonth.getMonth();
    const daysInMonth = getDaysInMonth(year, month);
    const firstDayOfMonth = getFirstDayOfMonth(year, month);
    
    const days = [];
    // Add empty cells for days before first day of month
    for (let i = 0; i < firstDayOfMonth; i++) {
      days.push(<div key={`empty-${i}`} className="h-9 w-9"></div>);
    }
    
    // Add cells for each day of the month
    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(year, month, day);
      const isStart = date.toDateString() === tempRange.start.toDateString();
      const isEnd = date.toDateString() === tempRange.end.toDateString();
      const isInRange = date >= tempRange.start && date <= tempRange.end;
      const isToday = date.toDateString() === new Date().toDateString();
      
      days.push(
        <button
          key={day}
          type="button"
          className={`h-9 w-9 rounded-full flex items-center justify-center text-sm
            ${isStart || isEnd ? 'bg-blue-600 text-white' : ''}
            ${isInRange && !isStart && !isEnd ? 'bg-blue-100 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400' : ''}
            ${isToday && !isStart && !isEnd ? 'border border-blue-600 dark:border-blue-400' : ''}
            ${!isInRange && !isToday ? 'hover:bg-gray-100 dark:hover:bg-gray-700' : ''}
          `}
          onClick={() => handleDateClick(date)}
        >
          {day}
        </button>
      );
    }
    
    return days;
  };

  const months = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ];

  return (
    <div className={`relative ${className}`}>
      <button
        type="button"
        className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        onClick={toggleDatepicker}
      >
        <Calendar className="h-4 w-4 mr-2" />
        <span>
          {formatDate(timeRange.start)} - {formatDate(timeRange.end)}
        </span>
      </button>

      {isOpen && (
        <div className="absolute mt-2 z-10 bg-white dark:bg-gray-800 shadow-lg rounded-md p-4 w-auto">
          <div className="flex items-center justify-between mb-4">
            <button
              type="button"
              className="p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700"
              onClick={previousMonth}
            >
              <ChevronLeft className="h-5 w-5 text-gray-600 dark:text-gray-400" />
            </button>
            <div className="text-sm font-medium text-gray-800 dark:text-white">
              {months[currentMonth.getMonth()]} {currentMonth.getFullYear()}
            </div>
            <button
              type="button"
              className="p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700"
              onClick={nextMonth}
            >
              <ChevronRight className="h-5 w-5 text-gray-600 dark:text-gray-400" />
            </button>
          </div>

          <div className="grid grid-cols-7 gap-1">
            {['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'].map((day) => (
              <div 
                key={day} 
                className="h-9 w-9 flex items-center justify-center text-xs font-medium text-gray-500 dark:text-gray-400"
              >
                {day}
              </div>
            ))}
            {renderCalendar()}
          </div>

          <div className="mt-4 flex items-center justify-between text-sm">
            <div className="text-gray-600 dark:text-gray-400">
              {selecting === 'start' ? 'Select start date' : 'Select end date'}
            </div>
            <button
              type="button"
              className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
              onClick={() => {
                onChange(tempRange);
                setIsOpen(false);
              }}
            >
              Apply
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default DateRangePicker;
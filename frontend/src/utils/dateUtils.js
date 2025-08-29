/**
 * Safely formats a date to Russian locale string
 * @param {string|Date} dateValue - Date value to format
 * @returns {string} Formatted date or fallback text
 */
export function formatDate(dateValue) {
  if (!dateValue) return 'Не указана';
  
  try {
    // Handle different date formats
    let dateStr = dateValue;
    if (typeof dateStr === 'string') {
      dateStr = dateStr.split('T')[0]; // Remove time part if present
    }
    
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) {
      console.error('Invalid date:', dateValue);
      return 'Неверная дата';
    }
    
    return date.toLocaleDateString('ru-RU');
  } catch (error) {
    console.error('Date parsing error:', dateValue, error);
    return 'Неверная дата';
  }
}

/**
 * Safely formats a date to Russian locale string, returns '-' for null/undefined
 * @param {string|Date} dateValue - Date value to format
 * @returns {string} Formatted date or '-' for empty values
 */
export function formatDateOptional(dateValue) {
  if (!dateValue) return '-';
  return formatDate(dateValue);
}
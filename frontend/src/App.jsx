import { useState, useEffect, useRef } from 'react';
import { Play, Pause, FolderOpen, Volume2, Maximize } from 'lucide-react';
// Wails জেনারেট করা বাইন্ডিংস ইমপোর্ট
import { SelectAndPlay, TogglePlay, SetVolume, Seek, GetPlaybackStatus, ResizeVideoWindow } from '../wailsjs/go/main/App';
import * as runtime from '@wailsapp/runtime'; // Wails Runtime

function App() {
  const [media, setMedia] = useState({ title: "No Video Loaded", duration: 0 });
  const [status, setStatus] = useState({ current: 0, total: 0 });
  const [isPlaying, setIsPlaying] = useState(false);
  const [volume, setVolume] = useState(80);
  const progressRef = useRef<HTMLDivElement>(null);

  // ১. মিডিয়া ওপেন করা
  const handleOpen = async () => {
    try {
      const info = await SelectAndPlay();
      setMedia(info);
      setIsPlaying(true);
    } catch (e) {
      console.error(e);
    }
  };

  // ২. প্লে/পজ
  const handleTogglePlay = async () => {
    const playing = await TogglePlay();
    setIsPlaying(playing);
  };

  // ৩. ভলিউম
  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const vol = parseInt(e.target.value);
    setVolume(vol);
    SetVolume(vol);
  };

  // ৪. সিক (Seeking) - প্রগ্রেস বারে ক্লিক করলে
  const handleSeek = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!progressRef.current || media.duration === 0) return;
    const rect = progressRef.current.getBoundingClientRect();
    const pos = (e.clientX - rect.left) / rect.width;
    Seek(pos);
  };

  // ৫. প্লেব্যাক স্ট্যাটাস আপডেট করার পোলিং (Polling)
  useEffect(() => {
    const interval = setInterval(async () => {
      if (isPlaying) {
        const [current, total] = await GetPlaybackStatus();
        setStatus({ current, total });
      }
    }, 500); // প্রতি ৫০০ মিলিসেকেন্ডে আপডেট
    return () => clearInterval(interval);
  }, [isPlaying]);

  // ৬. উইন্ডো রিসাইজ হ্যান্ডেল করা (Go কে জানানো)
  useEffect(() => {
    const handleResize = () => {
      ResizeVideoWindow();
    };
    window.addEventListener('resize', handleResize);
    // শুরুতে একবার রিসাইজ
    ResizeVideoWindow();
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // সময় ফরম্যাট করা (ms -> mm:ss)
  const formatTime = (ms: number) => {
    const totalSeconds = Math.floor(ms / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds < 10 ? '0' : ''}${seconds}`;
  };

  const progressPercentage = status.total > 0 ? (status.current / status.total) * 100 : 0;

  return (
    <div className="h-screen bg-[#09090b] text-white flex flex-col font-sans select-none overflow-hidden">
      
      {/* Video Viewport Area (Top 80%) */}
      {/* Go code will render VLC here, below we just provide background */}
      <div className="flex-1 bg-black flex items-center justify-center">
        {!isPlaying && (
            <div className='text-zinc-700 text-center p-10'>
                <FolderOpen className='w-16 h-16 mx-auto mb-4'/>
                <p>Native VLC will render here when video is loaded.</p>
                <p className='text-xs mt-1'>Windows API will take over this area.</p>
            </div>
        )}
      </div>

      {/* Modern Control Bar (Bottom 100px) */}
      <div className="h-[100px] border-t border-zinc-800 bg-[#09090b] px-6 py-3 flex flex-col justify-between">
        
        {/* Playback Title & Time */}
        <div className="flex items-center justify-between text-xs text-zinc-500 mb-1">
          <p className="truncate max-w-[300px] text-zinc-200">{media.title}</p>
          <p>{formatTime(status.current)} / {formatTime(status.total)}</p>
        </div>

        {/* Custom Progress Bar */}
        <div 
          ref={progressRef}
          onClick={handleSeek}
          className="w-full h-1.5 bg-zinc-800 rounded-full cursor-pointer group mb-3 relative"
        >
          <div 
            style={{ width: `${progressPercentage}%` }} 
            className="h-full bg-white rounded-full group-hover:bg-sky-400 transition-all"
          />
          <div 
            style={{ left: `${progressPercentage}%` }}
            className='absolute -top-1 w-3 h-3 bg-white rounded-full opacity-0 group-hover:opacity-100 -translate-x-1/2'
          />
        </div>

        {/* Buttons and Volume */}
        <div className="flex items-center justify-between">
          <button onClick={handleOpen} className="p-2 hover:bg-zinc-800 rounded-lg transition text-zinc-400 hover:text-white">
            <FolderOpen size={20} />
          </button>
          
          <button 
            onClick={handleTogglePlay}
            className="p-3 bg-white rounded-full text-black hover:scale-105 transition active:scale-95"
          >
            {isPlaying ? <Pause className='fill-black' size={24} /> : <Play className='fill-black' size={24} />}
          </button>

          <div className="flex items-center gap-3 text-zinc-400">
            <Volume2 size={18} />
            <input 
              type="range" 
              min="0" 
              max="100" 
              value={volume} 
              onChange={handleVolumeChange}
              className="w-24 h-1 accent-white cursor-pointer"
            />
             <button onClick={() => runtime.WindowToggleMaximise()} className="p-1 hover:bg-zinc-800 rounded transition text-zinc-400 hover:text-white">
                <Maximize size={16}/>
             </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
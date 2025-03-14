import { Routes } from '@angular/router';
import { TopicsComponent } from './topics/topics.component';

export const routes: Routes = [
    { path: '', redirectTo: 'topics', pathMatch: 'full' },
    { path: 'topics', component: TopicsComponent }
];
